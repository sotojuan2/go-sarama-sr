# Random Shoe Data Generation Logic Design

## Overview
This document defines the logic and constraints for generating realistic random data for the `Shoe` struct, ensuring compliance with the Protobuf schema and real-world plausibility.

## Shoe Struct Analysis

Based on the Protobuf schema (`shoe.proto`) and generated Go code, the `Shoe` struct contains:

```go
type Shoe struct {
    Id        int64   // Unique identifier for the shoe
    Brand     string  // Brand name of the shoe  
    Name      string  // Product name of the shoe
    SalePrice float64 // Sale price of the shoe
    Rating    float64 // Rating of the shoe (typically 1-5 scale)
}
```

## Field Generation Logic

### 1. Id (int64)
- **Type**: Unique identifier
- **Range**: 1 to 999,999,999 (9 digits max for realistic product IDs)
- **Logic**: Generate random positive integers within range
- **Constraints**: 
  - Must be positive (> 0)
  - Should be unique within a reasonable timeframe
  - Use Unix timestamp + random suffix for uniqueness

### 2. Brand (string)
- **Type**: Shoe brand name
- **Source**: Predefined list of realistic shoe brands
- **Logic**: Random selection from curated brand list
- **Constraints**:
  - Non-empty string
  - Realistic brand names only
  - Length: 2-20 characters
- **Brand Pool**:
  - Nike, Adidas, Puma, Reebok, New Balance, Converse, Vans, ASICS
  - Jordan, Under Armour, Skechers, Timberland, Dr. Martens, Clarks
  - Balenciaga, Gucci, Yeezy, Off-White, Supreme, Stone Island

### 3. Name (string)
- **Type**: Product/model name
- **Logic**: Generate realistic shoe model names using patterns
- **Constraints**:
  - Non-empty string
  - Length: 5-50 characters
  - Realistic naming patterns
- **Generation Patterns**:
  - `{Brand} {Model} {Version}` (e.g., "Air Max 90")
  - `{Model} {Type}` (e.g., "Chuck Taylor All Star")
  - `{Name} {Edition}` (e.g., "Boost 350 V2")
- **Components**:
  - Models: Air Max, Boost, Chuck Taylor, Stan Smith, Gazelle, Blazer, etc.
  - Types: High, Low, Mid, Classic, Retro, Pro, etc.
  - Versions: 90, 95, 97, 270, 350, V2, V3, etc.

### 4. SalePrice (float64)
- **Type**: Price in currency units (assuming USD)
- **Range**: $19.99 to $899.99 (realistic shoe price range)
- **Logic**: Generate prices with realistic distribution
- **Constraints**:
  - Must be positive (> 0)
  - Precision: 2 decimal places (cents)
  - Realistic price ranges by category
- **Price Ranges**:
  - Budget: $19.99 - $79.99 (40% probability)
  - Mid-range: $80.00 - $199.99 (35% probability)  
  - Premium: $200.00 - $399.99 (20% probability)
  - Luxury: $400.00 - $899.99 (5% probability)

### 5. Rating (float64)
- **Type**: Customer rating score
- **Range**: 1.0 to 5.0 (standard 5-star rating system)
- **Logic**: Generate ratings with realistic distribution (skewed toward higher ratings)
- **Constraints**:
  - Range: 1.0 ≤ rating ≤ 5.0
  - Precision: 1 decimal place (e.g., 4.2, 3.7)
  - Realistic distribution (most products rated 3.5-4.5)
- **Distribution**:
  - 1.0-2.4: 5% (poor products)
  - 2.5-3.4: 15% (below average)
  - 3.5-4.4: 60% (good products)
  - 4.5-5.0: 20% (excellent products)

## Implementation Strategy

### Option 1: Pure Go Standard Library
- Use `math/rand` for random number generation
- Implement custom logic for each field
- Create predefined slices for brands and model components
- **Pros**: No external dependencies, full control
- **Cons**: More code to maintain

### Option 2: Faker Library Integration
- Research Go faker libraries (e.g., `go-faker/faker`, `brianvoe/gofakeit`)
- Use faker for basic types, custom logic for shoe-specific fields
- **Pros**: Less custom code, proven randomization
- **Cons**: External dependency, may need customization

### Option 3: Hybrid Approach (Recommended)
- Use standard library for numeric fields (ID, price, rating)
- Custom logic for shoe-specific strings (brand, name)
- Keep it simple and focused on shoe domain

## Validation Requirements

Generated `Shoe` objects must satisfy:

1. **Schema Compliance**: All fields match Protobuf types
2. **Data Integrity**: No nil/empty required fields
3. **Realistic Values**: Values make sense in real-world context
4. **Range Constraints**: All numeric values within defined ranges
5. **String Quality**: Non-empty strings with realistic content

## Testing Strategy

Unit tests should verify:
- Field types match schema
- Values within expected ranges
- String fields are non-empty and realistic
- Distribution of values over multiple generations
- No duplicate IDs in reasonable sample size
- Performance with large generation volumes

## Output Example

```go
// Example generated Shoe objects:
{Id: 1692123456789, Brand: "Nike", Name: "Air Max 270", SalePrice: 129.99, Rating: 4.2}
{Id: 1692123456790, Brand: "Adidas", Name: "Boost 350 V2", SalePrice: 219.99, Rating: 4.7}
{Id: 1692123456791, Brand: "Converse", Name: "Chuck Taylor All Star High", SalePrice: 64.99, Rating: 4.1}
```

## Next Steps

1. Choose implementation approach (recommend Hybrid)
2. Implement field generation functions
3. Create main `GenerateRandomShoe()` function
4. Write comprehensive unit tests
5. Validate realistic output quality
