// Handlers of all command line commands.
package main

import (
	"context"
	"fmt"
	"io"
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
	if args.Start != nil {
		handleStart(args, *args.Start)
	} else if args.Stop != nil {
		handleStop(args, *args.Stop)
	} else if args.List != nil {
		handleList(args, *args.List)
	} else if args.Log != nil {
		handleLog(args, *args.Log)
	} else if args.Status != nil {
		handleStatus(args, *args.Status)
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
func handleStart(args args, cmd startCmd) {
	fmt.Println("Starting command:", strings.Join(cmd.Command, " "))

	client, close := createClient(args)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	req := teleportproto.Command{Command: cmd.Command}
	st, err := client.Start(ctx, &req)
	if err != nil {
		fatalError(err, "could not start a new command")
	}
	fmt.Println("Started job", st.Id.Uuid)
}

// Handles the "stop" command - kills the remote process and obtains its status
func handleStop(args args, cmd stopCmd) {
	fmt.Println("Stopping job", cmd.JobID)
	client, close := createClient(args)
	defer close()
	taskID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	ctx, cancel := defaultContext()
	defer cancel()
	st, err := client.Stop(ctx, &taskID)
	if err != nil {
		fatalError(err, "could not stop the job")
	}
	fmt.Println("Stopped job")
	printStatus(st)
}

// Handles the "list" command - list statuses of running jobs
func handleList(args args, cmd listCmd) {
	fmt.Println("Listing jobs")
	client, close := createClient(args)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	list, err := client.List(ctx, &empty.Empty{})
	if err != nil {
		fatalError(err, "could not list jobs")
	}
	for _, status := range list.Jobs {
		fmt.Println(separator)
		printStatus(status)
	}
}

// Handles the "sttus" taks - shows status of the remote job
func handleStatus(args args, cmd statusCmd) {
	fmt.Println("Showing status for job", cmd.JobID)
	client, close := createClient(args)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	taskID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	status, err := client.GetStatus(ctx, &taskID)
	if err != nil {
		fatalError(err, "could not get status for the job")
	}
	printStatus(status)
}

// Handles the "log" command - streams logs of the remote job
func handleLog(args args, cmd logCmd) {
	fmt.Println("Showing logs for job", cmd.JobID)
	client, close := createClient(args)
	defer close()
	ctx := context.Background()
	jobID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	stream, err := client.Logs(ctx, &jobID)
	if err != nil {
		fatalError(err, "could not get logs for the job")
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("== End of logs ==")
			return
		} else if err != nil {
			fatalError(err, "could not receive logs")
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
func printStatus(status *teleportproto.JobStatus) {
	fmt.Println("Job ID :", status.Id.Uuid)
	fmt.Println("Command:", strings.Join(status.Command.Command, " "))
	fmt.Println("Started:", status.Started.AsTime())
	fmt.Println("Logs   :", status.Logs)
	if status.Details != nil {
		switch details := status.Details.(type) {
		case *teleportproto.JobStatus_Stopped:
			fmt.Println("Stopped:", details.Stopped.Stopped.AsTime())
			fmt.Println("E. code:", details.Stopped.ErrorCode)
		case *teleportproto.JobStatus_Pending:
			fmt.Printf("CPU %%  : %.2f\n", details.Pending.CpuPerc)
			fmt.Printf("Memory : %.0f\n", details.Pending.Memory)
		}
	}
}
