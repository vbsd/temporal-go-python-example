import argparse
import asyncio
import uuid

from loguru import logger
from temporalio.client import Client
from temporalio.worker import Worker

import names


def uuid4():
    return str(uuid.uuid4())


async def book_car(user_id: str) -> str:
    await asyncio.sleep(1)
    booking_id = uuid4()
    logger.info(f"Booked a car: {booking_id=}, {user_id=}")
    return booking_id


async def book_hotel(user_id: str) -> str:
    await asyncio.sleep(1)
    booking_id = uuid4()
    logger.info(f"Booked a hotel: {booking_id=}, {user_id=}")
    return booking_id


async def book_flight(user_id: str) -> str:
    await asyncio.sleep(1)
    booking_id = uuid4()
    logger.info(f"Booked a flight: {booking_id=}, {user_id=}")
    return booking_id


async def cancel_car_booking(user_id: str, booking_id: str) -> None:
    await asyncio.sleep(1)
    logger.info(f"Canceled a car booking: {booking_id=}, {user_id=}")


async def cancel_hotel_booking(user_id: str, booking_id: str) -> None:
    await asyncio.sleep(1)
    logger.info(f"Canceled a hotel booking: {booking_id=}, {user_id=}")


async def cancel_flight_booking(user_id: str, booking_id: str) -> None:
    await asyncio.sleep(1)
    logger.info(f"Canceled a flight booking: {booking_id=}, {user_id=}")


async def main(namespace: str, host: str):
    """Runs a worker with booking activities."""

    client = await Client.connect(f"http://{host}", namespace=namespace)
    worker = Worker(
        client,
        task_queue=names.TASK_QUEUE_ACTIVITIES,
        activities={
            names.ACTIVITY_BOOK_CAR: book_car,
            names.ACTIVITY_BOOK_HOTEL: book_hotel,
            names.ACTIVITY_BOOK_FLIGHT: book_flight,
            names.ACTIVITY_CANCEL_CAR_BOOKING: cancel_car_booking,
            names.ACTIVITY_CANCEL_HOTEL_BOOKING: cancel_hotel_booking,
            names.ACTIVITY_CANCEL_FLIGHT_BOOKING: cancel_flight_booking,
        },
    )
    await worker.run()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Temporal activity worker")
    parser.add_argument(
        "--host",
        default="temporal:7233",
        type=str,
        required=False,
        help="Temporal host and port (default 'temporal:7233')",
    )
    parser.add_argument(
        "--namespace",
        default="default",
        type=str,
        required=False,
        help="Temporal namespace (default 'default')",
    )
    args = parser.parse_args()
    asyncio.run(main(args.namespace, args.host))
