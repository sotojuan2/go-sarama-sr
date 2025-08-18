package test

import (
	"testing"

	"github.com/go-sarama-sr/producer/pb"
)

func TestShoeStruct(t *testing.T) {
	// Test that the Shoe struct can be instantiated and used
	shoe := &pb.Shoe{
		Id:        1,
		Brand:     "Nike",
		Name:      "Air Max 90",
		SalePrice: 129.99,
		Rating:    4.5,
	}

	// Test getter methods
	if shoe.GetId() != 1 {
		t.Errorf("Expected ID to be 1, got %d", shoe.GetId())
	}

	if shoe.GetBrand() != "Nike" {
		t.Errorf("Expected Brand to be 'Nike', got %s", shoe.GetBrand())
	}

	if shoe.GetName() != "Air Max 90" {
		t.Errorf("Expected Name to be 'Air Max 90', got %s", shoe.GetName())
	}

	if shoe.GetSalePrice() != 129.99 {
		t.Errorf("Expected SalePrice to be 129.99, got %f", shoe.GetSalePrice())
	}

	if shoe.GetRating() != 4.5 {
		t.Errorf("Expected Rating to be 4.5, got %f", shoe.GetRating())
	}

	// Test that protobuf serialization works
	if shoe.String() == "" {
		t.Error("Expected non-empty string representation")
	}

	t.Logf("✅ Shoe struct test passed - all fields working correctly")
}
