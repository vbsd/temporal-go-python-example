package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) registerBookingActivity(name string) {
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, userId string) (string, error) {
			return "", nil
		},
		activity.RegisterOptions{Name: name})
}

func (s *UnitTestSuite) registerCancellationActivity(name string) {
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, carBookingId string) error {
			return nil
		},
		activity.RegisterOptions{Name: name})
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	// We have to register these activities just so that TestSuite knows
	// their signatures. This would not be necessary if we were mocking
	// activities written in Go (and referenced by their function names),
	// but since we use activities outside of this Go project (written
	// in Python) and we refer to them by their name (as a simple string),
	// we have to do this.
	s.registerBookingActivity(ACTIVITY_BOOK_CAR)
	s.registerBookingActivity(ACTIVITY_BOOK_HOTEL)
	s.registerBookingActivity(ACTIVITY_BOOK_FLIGHT)

	s.registerCancellationActivity(ACTIVITY_CANCEL_CAR_BOOKING)
	s.registerCancellationActivity(ACTIVITY_CANCEL_HOTEL_BOOKING)
	s.registerCancellationActivity(ACTIVITY_CANCEL_FLIGHT_BOOKING)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	// Asserts that all mocked activities were called as expected.
	// If some of the activities that were mocked are not called
	// during the execution, this will fail the test.
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) mockActivityF0(name string, err error) {
	s.env.OnActivity(name, mock.Anything, mock.Anything).Return(err)
}

func (s *UnitTestSuite) mockActivityF1(name string, result string, err error) {
	s.env.OnActivity(name, mock.Anything, mock.Anything).Return(result, err)
}

func (s *UnitTestSuite) TestSuccess() {
	expectedResult := Trip{
		CarBookingId:    "car-booking-id",
		HotelBookingId:  "hotel-booking-id",
		FlightBookingId: "flight-booking-id",
	}
	s.mockActivityF1(ACTIVITY_BOOK_CAR, expectedResult.CarBookingId, nil)
	s.mockActivityF1(ACTIVITY_BOOK_HOTEL, expectedResult.HotelBookingId, nil)
	s.mockActivityF1(ACTIVITY_BOOK_FLIGHT, expectedResult.FlightBookingId, nil)

	s.env.ExecuteWorkflow(BookTrip, "user-id-success-test")
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result Trip
	s.env.GetWorkflowResult(&result)
	s.Equal(expectedResult, result)
}

func (s *UnitTestSuite) TestAllCompensated() {
	s.mockActivityF1(ACTIVITY_BOOK_CAR, "", errors.New(""))
	s.mockActivityF1(ACTIVITY_BOOK_HOTEL, "", errors.New(""))
	s.mockActivityF1(ACTIVITY_BOOK_FLIGHT, "", errors.New(""))
	s.mockActivityF0(ACTIVITY_CANCEL_CAR_BOOKING, nil)
	s.mockActivityF0(ACTIVITY_CANCEL_HOTEL_BOOKING, nil)
	s.mockActivityF0(ACTIVITY_CANCEL_FLIGHT_BOOKING, nil)

	s.env.ExecuteWorkflow(BookTrip, "user-id-all-compensated-test")

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
