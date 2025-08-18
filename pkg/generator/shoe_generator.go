package generator

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	pb "github.com/go-sarama-sr/producer/pb"
)

// ShoeGenerator provides functionality to generate random Shoe objects
type ShoeGenerator struct {
	rng *rand.Rand
}

// NewShoeGenerator creates a new instance of ShoeGenerator with a seeded random number generator
func NewShoeGenerator() *ShoeGenerator {
	return &ShoeGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Predefined data for realistic generation
var (
	// shoeBrands contains a curated list of realistic shoe brands
	shoeBrands = []string{
		"Nike", "Adidas", "Puma", "Reebok", "New Balance", "Converse", "Vans", "ASICS",
		"Jordan", "Under Armour", "Skechers", "Timberland", "Dr. Martens", "Clarks",
		"Balenciaga", "Gucci", "Yeezy", "Off-White", "Supreme", "Stone Island",
	}

	// shoeModels contains model name components for realistic shoe names
	shoeModels = []string{
		"Air Max", "Boost", "Chuck Taylor", "Stan Smith", "Gazelle", "Blazer",
		"Cortez", "Superstar", "Classic", "Runner", "Trainer", "Court", "Forum",
		"Gel", "Free", "React", "Zoom", "Dunk", "Force", "Huarache",
	}

	// shoeTypes contains type/edition components for shoe names
	shoeTypes = []string{
		"High", "Low", "Mid", "Classic", "Retro", "Pro", "Elite", "Premium",
		"Vintage", "Modern", "Sport", "Lifestyle", "Performance", "Street",
		"All Star", "Essential", "Original", "V2", "V3", "Edition",
	}

	// shoeVersions contains version numbers and identifiers
	shoeVersions = []string{
		"90", "95", "97", "270", "350", "720", "1", "2", "3", "4", "5",
		"2024", "2025", "LX", "SE", "QS", "OG", "NRG", "SP",
	}
)

// GenerateId creates a unique shoe ID using timestamp and random suffix
func (sg *ShoeGenerator) GenerateId() int64 {
	// Use current timestamp (in microseconds) + random 3-digit suffix for better uniqueness
	timestamp := time.Now().UnixMicro()
	randomSuffix := sg.rng.Intn(1000) // 0-999
	
	// Combine timestamp with random suffix to create unique ID
	// Using microseconds provides much better uniqueness than seconds
	return timestamp*1000 + int64(randomSuffix)
}

// GenerateBrand selects a random brand from the predefined list
func (sg *ShoeGenerator) GenerateBrand() string {
	return shoeBrands[sg.rng.Intn(len(shoeBrands))]
}

// GenerateName creates a realistic shoe name using pattern-based generation
func (sg *ShoeGenerator) GenerateName() string {
	// Choose generation pattern randomly
	patterns := []func() string{
		// Pattern 1: Model + Version (e.g., "Air Max 90")
		func() string {
			model := shoeModels[sg.rng.Intn(len(shoeModels))]
			version := shoeVersions[sg.rng.Intn(len(shoeVersions))]
			return fmt.Sprintf("%s %s", model, version)
		},
		// Pattern 2: Model + Type (e.g., "Chuck Taylor All Star")
		func() string {
			model := shoeModels[sg.rng.Intn(len(shoeModels))]
			shoeType := shoeTypes[sg.rng.Intn(len(shoeTypes))]
			return fmt.Sprintf("%s %s", model, shoeType)
		},
		// Pattern 3: Model + Type + Version (e.g., "Boost 350 V2")
		func() string {
			model := shoeModels[sg.rng.Intn(len(shoeModels))]
			version := shoeVersions[sg.rng.Intn(len(shoeVersions))]
			shoeType := shoeTypes[sg.rng.Intn(len(shoeTypes))]
			return fmt.Sprintf("%s %s %s", model, version, shoeType)
		},
	}

	// Select random pattern and generate name
	pattern := patterns[sg.rng.Intn(len(patterns))]
	return pattern()
}

// GenerateSalePrice creates a realistic price with weighted distribution
func (sg *ShoeGenerator) GenerateSalePrice() float64 {
	// Define price tiers with probabilities
	// Budget: 40%, Mid-range: 35%, Premium: 20%, Luxury: 5%
	randValue := sg.rng.Float64()

	var minPrice, maxPrice float64

	switch {
	case randValue < 0.40: // Budget tier (40%)
		minPrice, maxPrice = 19.99, 79.99
	case randValue < 0.75: // Mid-range tier (35%)
		minPrice, maxPrice = 80.00, 199.99
	case randValue < 0.95: // Premium tier (20%)
		minPrice, maxPrice = 200.00, 399.99
	default: // Luxury tier (5%)
		minPrice, maxPrice = 400.00, 899.99
	}

	// Generate random price within the selected tier
	price := minPrice + sg.rng.Float64()*(maxPrice-minPrice)
	
	// Round to 2 decimal places (cents precision)
	return math.Round(price*100) / 100
}

// GenerateRating creates a realistic rating with distribution skewed toward higher values
func (sg *ShoeGenerator) GenerateRating() float64 {
	// Define rating ranges with probabilities
	// Poor (1.0-2.4): 5%, Below Average (2.5-3.4): 15%
	// Good (3.5-4.4): 60%, Excellent (4.5-5.0): 20%
	randValue := sg.rng.Float64()

	var minRating, maxRating float64

	switch {
	case randValue < 0.05: // Poor (5%)
		minRating, maxRating = 1.0, 2.4
	case randValue < 0.20: // Below Average (15%)
		minRating, maxRating = 2.5, 3.4
	case randValue < 0.80: // Good (60%)
		minRating, maxRating = 3.5, 4.4
	default: // Excellent (20%)
		minRating, maxRating = 4.5, 5.0
	}

	// Generate random rating within the selected range
	rating := minRating + sg.rng.Float64()*(maxRating-minRating)
	
	// Round to 1 decimal place
	return math.Round(rating*10) / 10
}

// GenerateRandomShoe creates a complete Shoe object with realistic random data
func (sg *ShoeGenerator) GenerateRandomShoe() *pb.Shoe {
	return &pb.Shoe{
		Id:        sg.GenerateId(),
		Brand:     sg.GenerateBrand(),
		Name:      sg.GenerateName(),
		SalePrice: sg.GenerateSalePrice(),
		Rating:    sg.GenerateRating(),
	}
}

// GenerateRandomShoes creates multiple random Shoe objects
func (sg *ShoeGenerator) GenerateRandomShoes(count int) []*pb.Shoe {
	// Handle edge cases
	if count <= 0 {
		return []*pb.Shoe{}
	}
	
	shoes := make([]*pb.Shoe, count)
	for i := 0; i < count; i++ {
		shoes[i] = sg.GenerateRandomShoe()
		// Small delay to ensure unique timestamps for IDs
		time.Sleep(1 * time.Microsecond)
	}
	return shoes
}
