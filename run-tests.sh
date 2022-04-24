#!/bin/bash
set -eu

export TEMPORAL_NAMESPACE=test

echo "Waiting for Temporal cluster..."
until tctl namespace list &> /dev/null
do
    sleep 1
done

if ! [[ $(tctl n l | grep "^Name: test$") ]]
then
    echo "Registering test namespace."
    tctl --namespace test namespace register
fi

echo "Waiting for test namespace to be available..."
until tctl --namespace test namespace describe &> /dev/null
do
    sleep 1
done


printf "\n\nRUNNING WORKFLOW WORKER\n\n"
pushd /temporal/workflow-worker
go run . --namespace test --host temporal:7233 &
popd

printf "\n\nRUNNING PYTHON TESTS\n\n"
pushd /temporal/activity-worker
pytest workflow_tests.py
popd
