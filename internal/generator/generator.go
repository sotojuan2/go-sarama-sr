package generator

import (
	"math/rand"
	"time"

	pb "github.com/go-sarama-sr/producer/pb"
)

// ShoeGenerator generates random shoe data
type ShoeGenerator struct {
	rand *rand.Rand
}

// NewShoeGenerator creates a new shoe generator with random seed
func NewShoeGenerator() *ShoeGenerator {
	return &ShoeGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRandomShoe creates a random shoe with realistic data
func (g *ShoeGenerator) GenerateRandomShoe() *pb.Shoe {
	brands := []string{
		"Nike", "Adidas", "Puma", "Reebok", "Converse",
		"Vans", "New Balance", "ASICS", "Jordan", "Under Armour",
	}

	models := []string{
		"Air Max", "Air Force", "Boost", "Chuck Taylor", "Old Skool",
		"Stan Smith", "Gazelle", "Classic", "Runner", "Court",
		"Retro", "Pro", "Sport", "Training", "Casual",
	}

	suffixes := []string{
		"90", "270", "97", "95", "1", "2.0", "3.0", "Plus",
		"Ultra", "Lite", "Pro", "Elite", "Premium", "Essential",
	}

	// Generate random values
	id := g.rand.Int63n(999999) + 1 // 1 to 999999
	brand := brands[g.rand.Intn(len(brands))]
	model := models[g.rand.Intn(len(models))]
	suffix := suffixes[g.rand.Intn(len(suffixes))]
	name := model + " " + suffix

	// Price between $30 and $300
	salePrice := 30.0 + g.rand.Float64()*270.0

	// Rating between 2.5 and 5.0
	rating := 2.5 + g.rand.Float64()*2.5

	return &pb.Shoe{
		Id:        id,
		Brand:     brand,
		Name:      name,
		SalePrice: salePrice,
		Rating:    rating,
	}
}
