// Handlers of all command line commands.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/client/proto/teleportproto"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
)

const separator = "------------------------------------------------------------"

// Executes command using parsed arguments
func execute(args args) {
	client, close := createClient(args)
	defer close()
	if args.Start != nil {
		handleStart(args, client)
	} else if args.Stop != nil {
		handleStop(args, client)
	} else if args.List != nil {
		handleList(args, client)
	} else if args.Log != nil {
		handleLog(args, client)
	} else if args.Status != nil {
		handleStatus(args, client)
	}
}

// Creates instance of a client
// On failure stops the application
func createClient(args args) (teleportproto.RemoteExecutorClient, func()) {

	var opts []grpc.DialOption
	if args.Secret == "" {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	} else {
		perRPC := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: args.Secret,
		})}
		creds, err := credentials.NewClientTLSFromFile(args.CaPath, "x.test.example.com") // TODO
		if err != nil {
			fatalError(err, "failed to load credentials")
		}
		opts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(perRPC),
			grpc.WithTransportCredentials(creds),
		}
	}

	conn, err := grpc.NewClient(args.Address, opts...)
	if err != nil {
		fatalError(err, "did not connect to the server")
	}
	return teleportproto.NewRemoteExecutorClient(conn), func() { conn.Close() }
}

// Handles the "start" command - start remote process
func handleStart(args args, client teleportproto.RemoteExecutorClient) error {
	fmt.Println("Starting command:", strings.Join(args.Start.Command, " "))

	ctx, cancel := defaultContext()
	defer cancel()
	req := teleportproto.Command{Command: args.Start.Command}
	st, err := client.Start(ctx, &req)
	if err != nil {
		return fmt.Errorf("could not start a new command: %w", err)
	}
	fmt.Println("Started job", st.Id.Uuid)
	return nil
}

// Handles the "stop" command - kills the remote process and obtains its status
func handleStop(args args, client teleportproto.RemoteExecutorClient) error {
	fmt.Println("Stopping job", args.Stop.JobID)
	taskID := teleportproto.JobId{Uuid: string(args.Stop.JobID)}
	ctx, cancel := defaultContext()
	defer cancel()
	st, err := client.Stop(ctx, &taskID)
	if err != nil {
		return fmt.Errorf("could not stop the job: %w", err)
	}
	fmt.Println("Stopped job")
	printStatus(st, os.Stdout)
	return nil
}

// Handles the "list" command - list statuses of running jobs
func handleList(args args, client teleportproto.RemoteExecutorClient) error {
	fmt.Println("Listing jobs")
	ctx, cancel := defaultContext()
	defer cancel()
	list, err := client.List(ctx, &empty.Empty{})
	if err != nil {
		return fmt.Errorf("could not list jobs: %w", err)
	}
	for _, status := range list.Jobs {
		fmt.Println(separator)
		printStatus(status, os.Stdout)
	}
	return nil
}

// Handles the "status" command - shows status of the remote job
func handleStatus(args args, client teleportproto.RemoteExecutorClient) error {
	fmt.Println("Showing status for job", args.Status.JobID)
	ctx, cancel := defaultContext()
	defer cancel()
	taskID := teleportproto.JobId{Uuid: string(args.Status.JobID)}
	status, err := client.GetStatus(ctx, &taskID)
	if err != nil {
		return fmt.Errorf("could not get status for the job: %w", err)
	}
	printStatus(status, os.Stdout)
	return nil
}

// Handles the "log" command - streams logs of the remote job
func handleLog(args args, client teleportproto.RemoteExecutorClient) error {
	fmt.Println("Showing logs for job", args.Log.JobID)
	ctx := context.Background()
	jobID := teleportproto.JobId{Uuid: string(args.Log.JobID)}
	stream, err := client.Logs(ctx, &jobID)
	if err != nil {
		return fmt.Errorf("could not get logs for the job: %w", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("== End of logs ==")
			return nil
		} else if err != nil {
			return fmt.Errorf("could not receive logs: %w", err)
		}
		if resp.Text != "" {
			stderr := false
			if resp.Src == teleportproto.LogSource_LS_STDERR {
				stderr = true
			}
			printLog(resp.Text, resp.Timestamp.AsTime(), stderr)
		}
	}
}

// Most request should complete in 1 second
func defaultContext() (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	return ctx, cancel
}

// Prints status of the remove job.
func printStatus(status *teleportproto.JobStatus, w io.Writer) {
	fmt.Fprintf(w, "Job ID : %s\n", status.Id.Uuid)
	fmt.Fprintf(w, "Command: %s\n", strings.Join(status.Command.Command, " "))
	fmt.Fprintf(w, "Started: %s\n", status.Started.AsTime())
	fmt.Fprintf(w, "Logs   : %d\n", status.Logs)
	if status.Details != nil {
		switch details := status.Details.(type) {
		case *teleportproto.JobStatus_Stopped:
			fmt.Fprintf(w, "Stopped: %s\n", details.Stopped.Stopped.AsTime())
			fmt.Fprintf(w, "E. code: %d\n", details.Stopped.ErrorCode)
		case *teleportproto.JobStatus_Pending:
			fmt.Fprintf(w, "CPU %%  : %.2f\n", details.Pending.CpuPerc)
			fmt.Fprintf(w, "Memory : %.0f\n", details.Pending.Memory)
		}
	}
}
