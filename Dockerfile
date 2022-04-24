FROM ubuntu:22.04

RUN apt-get update
RUN apt-get install python3.10 python-is-python3 python3-pip golang -y
RUN pip3 install poetry
RUN apt-get install git -y

# Install tctl
RUN mkdir /build-tctl
RUN git clone https://github.com/temporalio/tctl.git /build-tctl
WORKDIR /build-tctl
# Checking out https://github.com/temporalio/tctl/releases/tag/v1.16.1
RUN git checkout 681f0e458b3ac54201c40abfc800852d97eca8f3
RUN make
RUN cp ./tctl /usr/bin/tctl
RUN rm -r /build-tctl

RUN mkdir /temporal

# Install Go dependencies for workflow worker
COPY ./workflow-worker /temporal/workflow-worker
WORKDIR /temporal/workflow-worker
RUN go build

# Install poetry dependencies for activity worker
COPY ./activity-worker /temporal/activity-worker
WORKDIR /temporal/activity-worker
RUN poetry config virtualenvs.create false
RUN poetry install

WORKDIR /temporal
