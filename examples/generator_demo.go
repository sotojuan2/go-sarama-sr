package main

import (
	"fmt"
	"strings"

	"github.com/go-sarama-sr/producer/pkg/generator"
)

func main() {
	// Create a new shoe generator
	shoeGen := generator.NewShoeGenerator()

	fmt.Println("=== Random Shoe Generation Examples ===\n")

	// Generate and display individual field examples
	fmt.Println("Individual Field Examples:")
	fmt.Printf("Random ID: %d\n", shoeGen.GenerateId())
	fmt.Printf("Random Brand: %s\n", shoeGen.GenerateBrand())
	fmt.Printf("Random Name: %s\n", shoeGen.GenerateName())
	fmt.Printf("Random Price: $%.2f\n", shoeGen.GenerateSalePrice())
	fmt.Printf("Random Rating: %.1f\n", shoeGen.GenerateRating())

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Generate complete shoe objects
	fmt.Println("\nComplete Shoe Objects:")
	for i := 0; i < 5; i++ {
		shoe := shoeGen.GenerateRandomShoe()
		fmt.Printf("Shoe %d:\n", i+1)
		fmt.Printf("  ID: %d\n", shoe.Id)
		fmt.Printf("  Brand: %s\n", shoe.Brand)
		fmt.Printf("  Name: %s\n", shoe.Name)
		fmt.Printf("  Price: $%.2f\n", shoe.SalePrice)
		fmt.Printf("  Rating: %.1f/5.0\n", shoe.Rating)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 50))

	// Generate multiple shoes at once
	fmt.Println("\nBatch Generation (10 shoes):")
	shoes := shoeGen.GenerateRandomShoes(10)
	for i, shoe := range shoes {
		fmt.Printf("%d. %s %s - $%.2f (%.1f★)\n",
			i+1, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)
	}

	fmt.Println("\n=== Generation Complete ===")
}
