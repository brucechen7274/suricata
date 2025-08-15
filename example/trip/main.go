package main

import (
	"context"
	"fmt"

	"github.com/ostafen/suricata/example/trip/travel"
	"github.com/ostafen/suricata/runtime/ollama"
)

func main() {
	invoker := ollama.NewInvoker(
		ollama.DefaultBaseURL,
		"granite3.3:8b",
		ollama.Options{
			NumCtx:      131072,
			Temperature: 0.1,
		},
	)

	itineraryAgent := travel.NewItineraryAgent(invoker)
	flightAgent := travel.NewFlightAgent(invoker, &flightTools{})
	hotelAgent := travel.NewHotelAgent(invoker, &hotelTools{})

	reply, err := itineraryAgent.ExtractInfo(context.Background(), &travel.ItineraryRequest{
		Request: `Plan a trip from Milan to Catania (Italy) for a few days (3-5) in middle August.`,
	})
	if err != nil {
		panic(err)
	}

	_, err = flightAgent.SearchFlights(context.Background(), &travel.FlightRequest{
		From:      reply.From,
		To:        reply.To,
		Date:      reply.StartDate,
		RoundTrip: true,
	})
	if err != nil {
		panic(err)
	}

	hotelReply, err := hotelAgent.BookHotel(context.Background(), &travel.HotelRequest{
		Location:     reply.To,
		CheckinDate:  reply.StartDate,
		CheckoutDate: reply.EndDate,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(hotelReply)
}

type flightTools struct{}

func (tools *flightTools) FindFlights(ctx context.Context, in *travel.FlightRequest) (*travel.FlightReply, error) {
	fmt.Printf("[FLIGHT AGENT] Searching flights options from %s (%s) to %s (%s), round_trip: %t\n", in.From.City, in.From.Country, in.To.City, in.To.Country, in.RoundTrip)

	return &travel.FlightReply{
		Flights: []travel.Flight{
			{
				Id:        "1",
				Cost:      100,
				RoundTrip: true,
			},
			{
				Id:        "2",
				Cost:      50,
				RoundTrip: false,
			},
			{
				Id:        "3",
				Cost:      80,
				RoundTrip: true,
			},
		},
	}, nil
}

func (tools *flightTools) BookFlight(ctx context.Context, in *travel.BookFlightRequest) (*travel.BookFlightReply, error) {
	fmt.Printf("[FLIGHT AGENT] Book flight with id %d\n", in.Id)

	return &travel.BookFlightReply{
		Booked: true,
	}, nil
}

type hotelTools struct{}

func (tools *hotelTools) FindHotels(ctx context.Context, in *travel.FindHotelRequest) (*travel.FindHotelReply, error) {
	fmt.Printf("[HOTEL AGENT] Searching hotels options in %s (%s)\n", in.Location.City, in.Location.Country)

	return &travel.FindHotelReply{
		Hotels: []travel.Hotel{
			{
				Name: "Hotel 1",
			},
			{
				Name: "Hotel 2",
			},
		},
	}, nil
}

func (tools *hotelTools) BookHotel(ctx context.Context, in *travel.BookHotelRequest) (*travel.BookHotelReply, error) {
	fmt.Printf("[HOTEL AGENT] Booking hotel %s from %s to %s\n", in.Name, in.CheckinDate, in.CheckoutDate)

	return &travel.BookHotelReply{Booked: true}, nil
}
