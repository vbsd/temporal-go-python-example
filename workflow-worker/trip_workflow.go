package main

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Trip struct {
	CarBookingId    string
	HotelBookingId  string
	FlightBookingId string
}

func BookTrip(ctx workflow.Context, userId string) (*Trip, error) {
	ao := workflow.ActivityOptions{
		TaskQueue:           TASK_QUEUE_ACTIVITIES,
		StartToCloseTimeout: 3 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("BookTrip workflow started", "userId", userId)

	// We run booking activities in parallel. ExecuteActivity returns a future
	// and does not block.
	carFuture := workflow.ExecuteActivity(ctx, ACTIVITY_BOOK_CAR, userId)
	hotelFuture := workflow.ExecuteActivity(ctx, ACTIVITY_BOOK_HOTEL, userId)
	flightFuture := workflow.ExecuteActivity(ctx, ACTIVITY_BOOK_FLIGHT, userId)

	var carBookingId string
	var hotelBookingId string
	var flightBookingId string
	carErr := carFuture.Get(ctx, &carBookingId)
	hotelErr := hotelFuture.Get(ctx, &hotelBookingId)
	flightErr := flightFuture.Get(ctx, &flightBookingId)

	success := (carErr == nil) && (hotelErr == nil) && (flightErr == nil)
	if success {
		result := Trip{
			CarBookingId:    carBookingId,
			HotelBookingId:  hotelBookingId,
			FlightBookingId: flightBookingId,
		}
		logger.Info("BookTrip workflow completed.", "result", result)
		return &result, nil
	}

	// If any one of bookings failed, we call cancellation activities.
	// Note that we call those activities for all bookings, not just the ones
	// that succeeded - it's the activity's responsibility to figure out if the
	// booking was created and deal with it.
	carFuture = workflow.ExecuteActivity(ctx, ACTIVITY_CANCEL_CAR_BOOKING, userId, carBookingId)
	hotelFuture = workflow.ExecuteActivity(ctx, ACTIVITY_CANCEL_HOTEL_BOOKING, userId, hotelBookingId)
	flightFuture = workflow.ExecuteActivity(ctx, ACTIVITY_CANCEL_FLIGHT_BOOKING, userId, flightBookingId)

	carErr = carFuture.Get(ctx, nil)
	hotelErr = hotelFuture.Get(ctx, nil)
	flightErr = flightFuture.Get(ctx, nil)

	if carErr == nil && hotelErr == nil && flightErr == nil {
		return nil, errors.New("Failed to complete bookings. All bookings canceled.")
	} else {
		// These cancellations can fail because we're using a retry policy
		// that limits retries. However, we could also retry forever or spin out
		// a different workflow that deals with failing cancellations.
		return nil, errors.New("Failed to complete bookings. Failed to cancel bookings.")
	}
}
