# go-teleport

Secure and scalable remote command executor and manager

g`o-teleport` is a project created with intention to be a sample of my skills. 
It can also be used as a solid project template for golang-based secure and efficient servers. 
It has the MIT license and you are free to use parts of the code in your projects.

# Challenge

This project is based on Teleport System Engineer interview task as defined [here](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge-1.md).

Create secure and scalable server and command line client with the capability to:

- Remotely start a task on the server. The task is any bash-like command.
- Query the task status.
- Connect to the server and obtain logs (stout and stderr) - both past outputs and a live stream of logs being generated by the process.
- Stop the task.
- List all pending tasks.

In addition to the functional requirements there are also several quality requirements:

- Create automated tests of the API.
- Limit task resources to make sure that processes run by one user do not affect other users.
- Provide a basic authorization mechanism.
- Use reproducible build system.
- Containerize the application (server).

# Design

Full design document can be found [here](./docs/design.md).

It contains discussion on the most important design choices and a brief graphical presentation of the communication process.

# Build and run

To build the application using docker:

```
docker build -t teleport .
```

You can also build it locally using the standard golang tools if you install few dependencies as defined in the [Dockerfile](./Dockerfile).
To run the application:

```
docker run -it --rm teleport /bin/bash
```

Then you can use commands `server` ad `client` to run the server and client apps. For help run:

```
server --help
client --help
```
