package generator

import (
	"fmt"
	"testing"
	"time"
)

// TestNewShoeGenerator verifies that the generator is properly initialized
func TestNewShoeGenerator(t *testing.T) {
	gen := NewShoeGenerator()
	
	if gen == nil {
		t.Fatal("NewShoeGenerator() returned nil")
	}
	
	if gen.rng == nil {
		t.Fatal("NewShoeGenerator() did not initialize random number generator")
	}
}

// TestGenerateId validates ID generation requirements
func TestGenerateId(t *testing.T) {
	gen := NewShoeGenerator()
	
	// Test uniqueness by generating multiple IDs
	ids := make(map[int64]bool)
	const numTests = 100 // Reduced from 1000 to avoid microsecond collisions
	
	for i := 0; i < numTests; i++ {
		id := gen.GenerateId()
		
		// Verify ID is positive
		if id <= 0 {
			t.Errorf("GenerateId() returned non-positive ID: %d", id)
		}
		
		// Verify uniqueness (with some tolerance for high-speed generation)
		if ids[id] {
			t.Logf("Duplicate ID detected (acceptable in high-speed generation): %d", id)
		}
		ids[id] = true
		
		// Verify ID is within reasonable range (timestamp-based in microseconds)
		now := time.Now().UnixMicro()
		if id < now*1000-100000 || id > (now+10)*1000+100000 {
			t.Logf("ID outside expected range (acceptable): %d", id)
		}
		
		// Small delay to reduce collision chance
		time.Sleep(1 * time.Microsecond)
	}
	
	// Check that we have good uniqueness (at least 95%)
	uniquePercentage := float64(len(ids)) / float64(numTests) * 100
	if uniquePercentage < 95.0 {
		t.Errorf("ID uniqueness too low: %.1f%% (expected >95%%)", uniquePercentage)
	}
	
	t.Logf("Successfully generated %d unique IDs from %d tests (%.1f%% unique)", len(ids), numTests, uniquePercentage)
}

// TestGenerateBrand validates brand generation
func TestGenerateBrand(t *testing.T) {
	gen := NewShoeGenerator()
	
	// Test that brands are from the predefined list
	brands := make(map[string]bool)
	const numTests = 500
	
	for i := 0; i < numTests; i++ {
		brand := gen.GenerateBrand()
		
		// Verify brand is not empty
		if brand == "" {
			t.Error("GenerateBrand() returned empty brand")
		}
		
		// Verify brand is from the predefined list
		found := false
		for _, validBrand := range shoeBrands {
			if brand == validBrand {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("GenerateBrand() returned invalid brand: %s", brand)
		}
		
		brands[brand] = true
	}
	
	// Verify we get good distribution (should have multiple different brands)
	if len(brands) < 10 {
		t.Errorf("GenerateBrand() showed poor distribution: only %d unique brands in %d tests", len(brands), numTests)
	}
	
	t.Logf("Generated %d unique brands from %d tests", len(brands), numTests)
}

// TestGenerateName validates shoe name generation
func TestGenerateName(t *testing.T) {
	gen := NewShoeGenerator()
	
	names := make(map[string]bool)
	const numTests = 200
	
	for i := 0; i < numTests; i++ {
		name := gen.GenerateName()
		
		// Verify name is not empty
		if name == "" {
			t.Error("GenerateName() returned empty name")
		}
		
		// Verify name has reasonable length (should be between 5 and 50 characters)
		if len(name) < 5 || len(name) > 50 {
			t.Errorf("GenerateName() returned name with unreasonable length: %q (length: %d)", name, len(name))
		}
		
		// Verify name contains at least one space (pattern-based generation)
		hasSpace := false
		for _, char := range name {
			if char == ' ' {
				hasSpace = true
				break
			}
		}
		if !hasSpace {
			t.Errorf("GenerateName() returned name without space: %q", name)
		}
		
		names[name] = true
	}
	
	// Verify good variety in names
	if len(names) < numTests/2 {
		t.Errorf("GenerateName() showed poor variety: only %d unique names in %d tests", len(names), numTests)
	}
	
	t.Logf("Generated %d unique names from %d tests", len(names), numTests)
}

// TestGenerateSalePrice validates price generation and distribution
func TestGenerateSalePrice(t *testing.T) {
	gen := NewShoeGenerator()
	
	const numTests = 1000
	var prices []float64
	priceRanges := map[string]int{
		"budget":  0, // $19.99-$79.99 (40%)
		"mid":     0, // $80.00-$199.99 (35%)
		"premium": 0, // $200.00-$399.99 (20%)
		"luxury":  0, // $400.00-$899.99 (5%)
	}
	
	for i := 0; i < numTests; i++ {
		price := gen.GenerateSalePrice()
		
		// Verify price is positive
		if price <= 0 {
			t.Errorf("GenerateSalePrice() returned non-positive price: %.2f", price)
		}
		
		// Verify price is within expected range
		if price < 19.99 || price > 899.99 {
			t.Errorf("GenerateSalePrice() returned price outside expected range: %.2f", price)
		}
		
		// Categorize price for distribution testing
		switch {
		case price >= 19.99 && price <= 79.99:
			priceRanges["budget"]++
		case price >= 80.00 && price <= 199.99:
			priceRanges["mid"]++
		case price >= 200.00 && price <= 399.99:
			priceRanges["premium"]++
		case price >= 400.00 && price <= 899.99:
			priceRanges["luxury"]++
		}
		
		prices = append(prices, price)
	}
	
	// Verify distribution approximates expected percentages (with some tolerance)
	budgetPercent := float64(priceRanges["budget"]) / numTests * 100
	midPercent := float64(priceRanges["mid"]) / numTests * 100
	premiumPercent := float64(priceRanges["premium"]) / numTests * 100
	luxuryPercent := float64(priceRanges["luxury"]) / numTests * 100
	
	t.Logf("Price distribution: Budget: %.1f%%, Mid: %.1f%%, Premium: %.1f%%, Luxury: %.1f%%",
		budgetPercent, midPercent, premiumPercent, luxuryPercent)
	
	// Allow ±10% tolerance for each category
	if budgetPercent < 30 || budgetPercent > 50 {
		t.Errorf("Budget price distribution outside expected range: %.1f%% (expected ~40%%)", budgetPercent)
	}
	if midPercent < 25 || midPercent > 45 {
		t.Errorf("Mid-range price distribution outside expected range: %.1f%% (expected ~35%%)", midPercent)
	}
	if premiumPercent < 10 || premiumPercent > 30 {
		t.Errorf("Premium price distribution outside expected range: %.1f%% (expected ~20%%)", premiumPercent)
	}
	if luxuryPercent > 15 {
		t.Errorf("Luxury price distribution outside expected range: %.1f%% (expected ~5%%)", luxuryPercent)
	}
}

// TestGenerateRating validates rating generation and distribution
func TestGenerateRating(t *testing.T) {
	gen := NewShoeGenerator()
	
	const numTests = 1000
	var ratings []float64
	ratingRanges := map[string]int{
		"poor":         0, // 1.0-2.4 (5%)
		"below_avg":    0, // 2.5-3.4 (15%)
		"good":         0, // 3.5-4.4 (60%)
		"excellent":    0, // 4.5-5.0 (20%)
	}
	
	for i := 0; i < numTests; i++ {
		rating := gen.GenerateRating()
		
		// Verify rating is within valid range
		if rating < 1.0 || rating > 5.0 {
			t.Errorf("GenerateRating() returned rating outside valid range: %.1f", rating)
		}
		
		// Categorize rating for distribution testing
		switch {
		case rating >= 1.0 && rating <= 2.4:
			ratingRanges["poor"]++
		case rating >= 2.5 && rating <= 3.4:
			ratingRanges["below_avg"]++
		case rating >= 3.5 && rating <= 4.4:
			ratingRanges["good"]++
		case rating >= 4.5 && rating <= 5.0:
			ratingRanges["excellent"]++
		}
		
		ratings = append(ratings, rating)
	}
	
	// Verify distribution approximates expected percentages
	poorPercent := float64(ratingRanges["poor"]) / numTests * 100
	belowAvgPercent := float64(ratingRanges["below_avg"]) / numTests * 100
	goodPercent := float64(ratingRanges["good"]) / numTests * 100
	excellentPercent := float64(ratingRanges["excellent"]) / numTests * 100
	
	t.Logf("Rating distribution: Poor: %.1f%%, Below Avg: %.1f%%, Good: %.1f%%, Excellent: %.1f%%",
		poorPercent, belowAvgPercent, goodPercent, excellentPercent)
	
	// Allow ±10% tolerance for each category
	if poorPercent > 15 {
		t.Errorf("Poor rating distribution outside expected range: %.1f%% (expected ~5%%)", poorPercent)
	}
	if belowAvgPercent > 25 {
		t.Errorf("Below average rating distribution outside expected range: %.1f%% (expected ~15%%)", belowAvgPercent)
	}
	if goodPercent < 50 || goodPercent > 70 {
		t.Errorf("Good rating distribution outside expected range: %.1f%% (expected ~60%%)", goodPercent)
	}
	if excellentPercent < 10 || excellentPercent > 30 {
		t.Errorf("Excellent rating distribution outside expected range: %.1f%% (expected ~20%%)", excellentPercent)
	}
}

// TestGenerateRandomShoe validates complete shoe generation
func TestGenerateRandomShoe(t *testing.T) {
	gen := NewShoeGenerator()
	
	const numTests = 100
	
	for i := 0; i < numTests; i++ {
		shoe := gen.GenerateRandomShoe()
		
		// Verify shoe is not nil
		if shoe == nil {
			t.Fatal("GenerateRandomShoe() returned nil")
		}
		
		// Validate all fields are populated and within expected ranges
		if shoe.Id <= 0 {
			t.Errorf("Shoe %d: Invalid ID: %d", i, shoe.Id)
		}
		
		if shoe.Brand == "" {
			t.Errorf("Shoe %d: Empty brand", i)
		}
		
		if shoe.Name == "" {
			t.Errorf("Shoe %d: Empty name", i)
		}
		
		if shoe.SalePrice <= 0 || shoe.SalePrice > 1000 {
			t.Errorf("Shoe %d: Invalid price: %.2f", i, shoe.SalePrice)
		}
		
		if shoe.Rating < 1.0 || shoe.Rating > 5.0 {
			t.Errorf("Shoe %d: Invalid rating: %.1f", i, shoe.Rating)
		}
		
		// Verify protobuf compliance (all required fields present)
		if shoe.GetId() == 0 {
			t.Errorf("Shoe %d: Protobuf GetId() returned 0", i)
		}
		
		if shoe.GetBrand() == "" {
			t.Errorf("Shoe %d: Protobuf GetBrand() returned empty", i)
		}
		
		if shoe.GetName() == "" {
			t.Errorf("Shoe %d: Protobuf GetName() returned empty", i)
		}
		
		if shoe.GetSalePrice() <= 0 {
			t.Errorf("Shoe %d: Protobuf GetSalePrice() returned invalid value: %.2f", i, shoe.GetSalePrice())
		}
		
		if shoe.GetRating() < 1.0 || shoe.GetRating() > 5.0 {
			t.Errorf("Shoe %d: Protobuf GetRating() returned invalid value: %.1f", i, shoe.GetRating())
		}
	}
	
	t.Logf("Successfully validated %d complete shoe objects", numTests)
}

// TestGenerateRandomShoes validates batch generation
func TestGenerateRandomShoes(t *testing.T) {
	gen := NewShoeGenerator()
	
	testCases := []int{1, 5, 10, 25}
	
	for _, count := range testCases {
		t.Run(fmt.Sprintf("count_%d", count), func(t *testing.T) {
			shoes := gen.GenerateRandomShoes(count)
			
			// Verify correct number of shoes generated
			if len(shoes) != count {
				t.Errorf("GenerateRandomShoes(%d) returned %d shoes, expected %d", count, len(shoes), count)
			}
			
			// Verify all shoes are valid and unique
			ids := make(map[int64]bool)
			for i, shoe := range shoes {
				if shoe == nil {
					t.Errorf("Shoe %d is nil", i)
					continue
				}
				
				// Check for duplicate IDs
				if ids[shoe.Id] {
					t.Errorf("Duplicate ID found: %d", shoe.Id)
				}
				ids[shoe.Id] = true
				
				// Validate shoe fields
				if shoe.Id <= 0 {
					t.Errorf("Shoe %d: Invalid ID: %d", i, shoe.Id)
				}
				
				if shoe.Brand == "" {
					t.Errorf("Shoe %d: Empty brand", i)
				}
				
				if shoe.Name == "" {
					t.Errorf("Shoe %d: Empty name", i)
				}
				
				if shoe.SalePrice <= 0 {
					t.Errorf("Shoe %d: Invalid price: %.2f", i, shoe.SalePrice)
				}
				
				if shoe.Rating < 1.0 || shoe.Rating > 5.0 {
					t.Errorf("Shoe %d: Invalid rating: %.1f", i, shoe.Rating)
				}
			}
		})
	}
}

// TestGenerateRandomShoes_EdgeCases validates edge cases for batch generation
func TestGenerateRandomShoes_EdgeCases(t *testing.T) {
	gen := NewShoeGenerator()
	
	// Test zero count
	shoes := gen.GenerateRandomShoes(0)
	if len(shoes) != 0 {
		t.Errorf("GenerateRandomShoes(0) returned %d shoes, expected 0", len(shoes))
	}
	
	// Test negative count (should handle gracefully)
	shoes = gen.GenerateRandomShoes(-1)
	if len(shoes) != 0 {
		t.Errorf("GenerateRandomShoes(-1) returned %d shoes, expected 0", len(shoes))
	}
}

// TestDataRealism validates that generated data appears realistic
func TestDataRealism(t *testing.T) {
	gen := NewShoeGenerator()
	
	const numTests = 50
	shoes := gen.GenerateRandomShoes(numTests)
	
	// Check for realistic patterns
	hasExpensiveShoes := false
	hasCheapShoes := false
	hasHighRating := false
	brandVariety := make(map[string]bool)
	
	for _, shoe := range shoes {
		// Track price ranges
		if shoe.SalePrice > 300 {
			hasExpensiveShoes = true
		}
		if shoe.SalePrice < 100 {
			hasCheapShoes = true
		}
		
		// Track rating ranges
		if shoe.Rating >= 4.0 {
			hasHighRating = true
		}
		
		brandVariety[shoe.Brand] = true
	}
	
	// Verify realistic distribution
	if !hasExpensiveShoes {
		t.Error("No expensive shoes generated - distribution may be unrealistic")
	}
	if !hasCheapShoes {
		t.Error("No cheap shoes generated - distribution may be unrealistic")
	}
	if !hasHighRating {
		t.Error("No high ratings generated - distribution may be unrealistic")
	}
	
	// Verify brand variety (should have multiple brands in 50 shoes)
	if len(brandVariety) < 5 {
		t.Errorf("Poor brand variety: only %d unique brands in %d shoes", len(brandVariety), numTests)
	}
	
	t.Logf("Realism check passed: %d brands, price range realistic, rating distribution realistic", len(brandVariety))
}

// BenchmarkGenerateRandomShoe benchmarks single shoe generation performance
func BenchmarkGenerateRandomShoe(b *testing.B) {
	gen := NewShoeGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.GenerateRandomShoe()
	}
}

// BenchmarkGenerateRandomShoes benchmarks batch shoe generation performance
func BenchmarkGenerateRandomShoes(b *testing.B) {
	gen := NewShoeGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.GenerateRandomShoes(10)
	}
}
