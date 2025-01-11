package main

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/szymonwieloch/go-teleport/client/mocks"
	"github.com/szymonwieloch/go-teleport/client/proto/teleportproto"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStartCommand(t *testing.T) {
	ctr := gomock.NewController(t)
	client := mocks.NewMockRemoteExecutorClient(ctr)
	args := args{
		Start: &startCmd{
			Command: []string{"echo", "blah"},
		},
	}

	cmd := teleportproto.Command{Command: []string{"echo", "blah"}}
	client.EXPECT().Start(gomock.Any(), gomock.Eq(&cmd)).Return(&exampleJobStatus, nil)
	err := handleStart(args, client)
	assert.NoError(t, err)
}

func TestListCommand(t *testing.T) {
	ctr := gomock.NewController(t)
	client := mocks.NewMockRemoteExecutorClient(ctr)
	jobList := teleportproto.JobList{}
	client.EXPECT().List(gomock.Any(), gomock.Any()).Return(&jobList, nil)
	args := args{
		List: &listCmd{},
	}
	err := handleList(args, client)
	assert.NoError(t, err)
}

func TestStopCommand(t *testing.T) {
	ctr := gomock.NewController(t)
	client := mocks.NewMockRemoteExecutorClient(ctr)
	args := args{
		Stop: &stopCmd{
			JobID: exampleJobID,
		},
	}
	expectedArg := &teleportproto.JobId{Uuid: exampleJobID}
	client.EXPECT().Stop(gomock.Any(), gomock.Eq(expectedArg)).Return(&exampleJobStatus, nil)
	err := handleStop(args, client)
	assert.NoError(t, err)
}

func TestStatusCommand(t *testing.T) {
	ctr := gomock.NewController(t)
	client := mocks.NewMockRemoteExecutorClient(ctr)
	args := args{
		Status: &statusCmd{
			JobID: exampleJobID,
		},
	}
	expectedArg := &teleportproto.JobId{Uuid: exampleJobID}
	client.EXPECT().GetStatus(gomock.Any(), gomock.Eq(expectedArg)).Return(&exampleJobStatus, nil)
	err := handleStatus(args, client)
	assert.NoError(t, err)
}

func TestLogCommand(t *testing.T) {
	ctr := gomock.NewController(t)
	client := mocks.NewMockRemoteExecutorClient(ctr)
	args := args{
		Log: &logCmd{
			JobID: exampleJobID,
		},
	}
	expectedArg := &teleportproto.JobId{Uuid: exampleJobID}

	stream := mocks.NewMockServerStreamingClient[teleportproto.Log](ctr)
	client.EXPECT().Logs(gomock.Any(), gomock.Eq(expectedArg)).Return(stream, nil)
	stream.EXPECT().Recv().Return(nil, io.EOF)
	err := handleLog(args, client)
	assert.NoError(t, err)
}

const exampleJobID = "6067dc56-0856-45f8-a87b-dd9745d292e7"

var exampleJobStatus teleportproto.JobStatus = teleportproto.JobStatus{
	Id:   &teleportproto.JobId{Uuid: exampleJobID},
	Logs: 15,
	Started: timestamppb.New(time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)),
	Command: &teleportproto.Command{Command: []string{"echo", "blah"}},
	Details: &teleportproto.JobStatus_Pending{
		Pending: &teleportproto.PendingJobStatus{
			CpuPerc: 12.0,
			Memory:  12345678,
		},
	},
}

func TestPrintStatus(t *testing.T) {

	buf := strings.Builder{}
	printStatus(&exampleJobStatus, &buf)
	want := "Job ID : 6067dc56-0856-45f8-a87b-dd9745d292e7\nCommand: echo blah\nStarted: 2009-11-17 20:34:58.651387237 +0000 UTC\nLogs   : 15\nCPU %  : 12.00\nMemory : 12345678\n"
	assert.Equal(t, buf.String(), want)
}
