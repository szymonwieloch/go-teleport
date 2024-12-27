#!/bin/bash

# this script emulates a long running process that prints to both stdout and stderr

for i in {1..1000}
do
    echo "doing something important ${i}th time"
    sleep 1
    echo error number $i >&2
    sleep 1
done