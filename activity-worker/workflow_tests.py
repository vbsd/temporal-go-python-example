import os
import uuid
from contextlib import asynccontextmanager
from dataclasses import dataclass
from typing import Coroutine, Optional

import pytest
from temporalio.client import Client, WorkflowFailureError
from temporalio.worker import Worker

import names


@dataclass
class Activities:
    book_car: Optional[Coroutine] = None
    book_hotel: Optional[Coroutine] = None
    book_flight: Optional[Coroutine] = None
    cancel_car_booking: Optional[Coroutine] = None
    cancel_hotel_booking: Optional[Coroutine] = None
    cancel_flight_booking: Optional[Coroutine] = None


def return_result(result=None):
    """return_result returns a function that returns the given result."""

    async def f(*args, **kwargs):
        return result

    return f


def raise_exception(message=None):
    """raise_exception returns a function that raises an exception with the given message."""

    async def f(*args, **kwargs):
        raise Exception(message)

    return f


@asynccontextmanager
async def worker_client(activities: Activities):
    """Runs the activities worker and returns a client."""

    host = os.getenv("TEMPORAL_HOST", "temporal:7233")
    client = await Client.connect(f"http://{host}", namespace="test")
    async with Worker(
        client,
        task_queue=names.TASK_QUEUE_ACTIVITIES,
        activities={
            names.ACTIVITY_BOOK_CAR: activities.book_car
            or raise_exception("Activity not implemented"),
            names.ACTIVITY_BOOK_HOTEL: activities.book_hotel
            or raise_exception("Activity not implemented"),
            names.ACTIVITY_BOOK_FLIGHT: activities.book_flight
            or raise_exception("Activity not implemented"),
            names.ACTIVITY_CANCEL_CAR_BOOKING: activities.cancel_car_booking
            or raise_exception("Activity not implemented"),
            names.ACTIVITY_CANCEL_HOTEL_BOOKING: activities.cancel_hotel_booking
            or raise_exception("Activity not implemented"),
            names.ACTIVITY_CANCEL_FLIGHT_BOOKING: activities.cancel_flight_booking
            or raise_exception("Activity not implemented"),
        },
    ):
        yield client


@pytest.mark.asyncio
async def test_workflow_success():
    """Test: if the bookings succeed, the workflow must also succeed."""

    expected_result = {
        "CarBookingId": "car-booking-id",
        "HotelBookingId": "hotel-booking-id",
        "FlightBookingId": "flight-booking-id",
    }
    activities = Activities(
        book_car=return_result(expected_result["CarBookingId"]),
        book_hotel=return_result(expected_result["HotelBookingId"]),
        book_flight=return_result(expected_result["FlightBookingId"]),
    )
    async with worker_client(activities) as client:
        result = await client.execute_workflow(
            "book-trip",
            "user-id",
            id=f"workflow-success-{uuid.uuid4()}",
            task_queue=names.TASK_QUEUE_WORKFLOWS,
        )
        assert expected_result == result


@pytest.mark.asyncio
async def test_workflow_fails():
    """Test: if the bookings fail, the workflow should execute cancellations
    and return an error with a particular message."""
    activities = Activities(
        book_car=raise_exception(ValueError),
        book_hotel=raise_exception(ValueError),
        book_flight=raise_exception(ValueError),
        cancel_car_booking=return_result("car-booking-canceled"),
        cancel_hotel_booking=return_result("hotel-booking-canceled"),
        cancel_flight_booking=return_result("flight-booking-canceled"),
    )
    async with worker_client(activities) as client:
        error_message = None
        try:
            await client.execute_workflow(
                "book-trip",
                "user-id",
                id=f"workflow-failure-{uuid.uuid4()}",
                task_queue=names.TASK_QUEUE_WORKFLOWS,
            )
        except WorkflowFailureError as e:
            error_message = e.cause.message

        assert error_message == "Failed to complete bookings. All bookings canceled."
