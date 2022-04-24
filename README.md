# Temporal example with Go workflows and Python activities

This is a sandbox that allows you to play around with Temporal using a Go workflow worker and a Python activity worker. It implements a trip booking saga (read more about the [Saga pattern](https://microservices.io/patterns/data/saga.html)) similar to the one described in Temporal [Booking Saga Tutorial in PHP](https://docs.temporal.io/docs/php/booking-saga-tutorial/). It showcases how you might use workflow and activity workers written in different languages.

## Project struture

* `./docker-compose.yaml` runs Temporal server, including our workflow and activity workers. It exposes Temporal Web UI at [localhost:8088](http://localhost:8088). It also runs a `temporal-development` container that has Go and Python dependencies installed and code mounted. You can use this container to run Temporal CLI tool `tctl` commands, run workflows and activities, modify code, run tests. If you are using [VS Code](https://code.visualstudio.com/), you might find it useful to connect to this container with [Remote Containers extensions](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers).

* `./Dockerfile` - `temporal-development` image definition.

* `./workflow-worker` - Go implementation of a workflow worker implementing a trip booking saga.

* `./activity-worker` - Python implementation of activities that the booking saga workflow uses. This also contains `workflow_tests.py` - these are examples of live tests that run the booking workflow and provide different activity implementations to show how the workflow acts.

* `./run-tests.sh` - this is a script that runs live tests. It creates a `test` Temporal namespace, runs a test workflow worker and runs Python `workflow_tests.py`.

## Running Temporal server and workers

Running this example depends only on [Docker](https://docs.docker.com/) and [docker-compose](https://docs.docker.com/compose/). Run the containers:

```docker-compose up -d```

Then connect to the development container:

```
docker exec -it temporal-development /bin/bash
```

Check Temporal cluster health:

```
tctl cluster health
```

The output should look something like this:

```
temporal.api.workflowservice.v1.WorkflowService: SERVING
```

## Running the booking saga workflow

You can run the booking workflow with Temporal CLI tool, `tctl`:

```
tctl workflow run \
    --taskqueue booking-workflows \
    --workflow_type book-trip \
    --workflow_id YOUR-WORKFLOW-ID \
    --input '"YOUR-USER-ID"'

```

`--workflow_id` is optional. If you omit it, the ID will be generated randomly. `'"YOUR-USER-ID"'` is double-quoted because input values must be JSON values.

You can inspect the workflow events at [http://localhost:8088/namespaces/default/workflows]().

## Running tests

There are two kinds of tests in this repository: Go unit tests and Python live tests.

Go unit tests are at `./workflow-worker/trip_workflow_test.go`. They are light, they use Go Temporal SDK's `testsuite` and mock the activities. To run these tests:

```
cd /temporal/workflow-worker
go test
```

Lear more about Go Temporal testing in [Temporal Go SDK testing docs](https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go).

Python tests (`./activity-workflow/workflow_tests.py`) are more involved. They run the `book-trip` workflow on a real Temporal server but provide different activity implementations for each test and inspect the result. These tests require:

* A running Temporal server which has a namespace `test` created.
* A workflow worker running in that namespace.

You can create the `test` namespace, run the workflow worker and run the tests just by running a helper script:

```
/temporal/run-tests.sh
```

 Alternatively, you can do this by hand:

1. `tctl --namespace test namespace register`
2. `cd /temporal/workflow-worker && go run . --namespace test --host temporal:7233 &` (running as a background task)
3. `cd /temporal/activity-worker && pytest workflow_tests.py`

After the tests are run, you can inspect the workflows at [http://localhost:8088/namespaces/test/workflows](). You should see one workflow that succeeded and one forkflow that failed with a message `Failed to complete bookings. All bookings canceled.`.

## Using workers without Docker

If you want to use the examples without the Docker, you'll need:

* A running Temporal server ([how-to](https://docs.temporal.io/docs/clusters/quick-install)).
* Go installed ([how-to](https://go.dev/doc/install)).
* Python 3.10 and poetry installed ([how-to](https://python-poetry.org/docs/)).

To run the workflow worker:

```
cd ./workflow-worker
go build
./worker --namespace <default> --host <host:port>
```

To run the activity worker:

```
cd ./activity-worker
poetry install
poetry run python main.py --namespace <default> --host <host:port>
```