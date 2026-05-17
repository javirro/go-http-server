package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// shopStore is the central in-memory store for the entire application.
type shopStore struct {
	mu sync.RWMutex

	products   map[int64]*Product
	productSeq int64

	// collectionProducts maps collection ID → ordered list of product IDs.
	collections          map[int64]*Collection
	collectionProducts   map[int64][]int64
	collectionSeq        int64

	carts    map[string]*Cart
	cartSeq  int64

	orders    map[int64]*Order
	orderSeq  int64
}

var shop = newShopStore()

func newShopStore() *shopStore {
	s := &shopStore{
		products:           make(map[int64]*Product),
		collections:        make(map[int64]*Collection),
		collectionProducts: make(map[int64][]int64),
		carts:              make(map[string]*Cart),
		orders:             make(map[int64]*Order),
	}
	s.seed()
	return s
}

// nextProductID allocates the next product ID (must hold write lock).
func (s *shopStore) nextProductID() int64 {
	s.productSeq++
	return s.productSeq
}

func (s *shopStore) nextCollectionID() int64 {
	s.collectionSeq++
	return s.collectionSeq
}

func (s *shopStore) nextCartItemID() int64 {
	s.cartSeq++
	return s.cartSeq
}

func (s *shopStore) nextOrderID() int64 {
	s.orderSeq++
	return s.orderSeq
}

// cartToken generates a random hex token for a cart.
func cartToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// seed pre-populates the store with football team jersey products and collections.
func (s *shopStore) seed() {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	pub := now

	sizes := []string{"XS", "S", "M", "L", "XL", "XXL"}

	type productSeed struct {
		title       string
		handle      string
		bodyHTML    string
		vendor      string
		productType string
		tags        string
		price       string
		compareAt   string
		imageSrc    string
		imageAlt    string
	}

	seeds := []productSeed{
		{
			title:       "Camiseta Real Madrid Equipación Local 2024/25",
			handle:      "camiseta-real-madrid-local-2024-25",
			bodyHTML:    "<p>La camiseta oficial de equipación local del Real Madrid para la temporada 2024/25. Fabricada en tejido Aeroready de Adidas para máxima transpirabilidad durante el juego.</p>",
			vendor:      "Adidas",
			productType: "Camiseta",
			tags:        "real-madrid, local, la-liga, adidas, 2024-25",
			price:       "89.95",
			compareAt:   "99.95",
			imageSrc:    "https://cdn.football-store.example/products/real-madrid-local-2024-25/1.jpg",
			imageAlt:    "Camiseta Real Madrid Local 2024/25",
		},
		{
			title:       "Camiseta Real Madrid Equipación Visitante 2024/25",
			handle:      "camiseta-real-madrid-visitante-2024-25",
			bodyHTML:    "<p>Equipación visitante oficial del Real Madrid 2024/25 con el inconfundible diseño azul claro de Adidas.</p>",
			vendor:      "Adidas",
			productType: "Camiseta",
			tags:        "real-madrid, visitante, la-liga, adidas, 2024-25",
			price:       "89.95",
			compareAt:   "",
			imageSrc:    "https://cdn.football-store.example/products/real-madrid-visitante-2024-25/1.jpg",
			imageAlt:    "Camiseta Real Madrid Visitante 2024/25",
		},
		{
			title:       "Camiseta FC Barcelona Equipación Local 2024/25",
			handle:      "camiseta-barcelona-local-2024-25",
			bodyHTML:    "<p>La icónica camiseta blaugrana del FC Barcelona para la temporada 2024/25, diseñada por Nike con tecnología Dri-FIT ADV.</p>",
			vendor:      "Nike",
			productType: "Camiseta",
			tags:        "barcelona, local, la-liga, nike, 2024-25",
			price:       "94.95",
			compareAt:   "",
			imageSrc:    "https://cdn.football-store.example/products/barcelona-local-2024-25/1.jpg",
			imageAlt:    "Camiseta FC Barcelona Local 2024/25",
		},
		{
			title:       "Camiseta FC Barcelona Equipación Visitante 2024/25",
			handle:      "camiseta-barcelona-visitante-2024-25",
			bodyHTML:    "<p>Equipación visitante del FC Barcelona 2024/25. Diseño dorado elegante de Nike que rinde homenaje a la historia del club.</p>",
			vendor:      "Nike",
			productType: "Camiseta",
			tags:        "barcelona, visitante, la-liga, nike, 2024-25",
			price:       "94.95",
			compareAt:   "104.95",
			imageSrc:    "https://cdn.football-store.example/products/barcelona-visitante-2024-25/1.jpg",
			imageAlt:    "Camiseta FC Barcelona Visitante 2024/25",
		},
		{
			title:       "Camiseta Atlético de Madrid Equipación Local 2024/25",
			handle:      "camiseta-atletico-madrid-local-2024-25",
			bodyHTML:    "<p>La camiseta a rayas rojas y blancas del Atlético de Madrid para la temporada 2024/25. Fabricada por Nike con tecnología Dri-FIT.</p>",
			vendor:      "Nike",
			productType: "Camiseta",
			tags:        "atletico-madrid, local, la-liga, nike, 2024-25",
			price:       "84.95",
			compareAt:   "",
			imageSrc:    "https://cdn.football-store.example/products/atletico-madrid-local-2024-25/1.jpg",
			imageAlt:    "Camiseta Atlético de Madrid Local 2024/25",
		},
		{
			title:       "Camiseta Selección Española Equipación Local UEFA Euro 2024",
			handle:      "camiseta-espana-local-euro-2024",
			bodyHTML:    "<p>La camiseta oficial de la Selección Española con la que conquistaron la Eurocopa 2024. Diseño rojo pasión de Adidas.</p>",
			vendor:      "Adidas",
			productType: "Camiseta",
			tags:        "espana, seleccion, euro-2024, adidas, local",
			price:       "99.95",
			compareAt:   "109.95",
			imageSrc:    "https://cdn.football-store.example/products/espana-local-euro-2024/1.jpg",
			imageAlt:    "Camiseta Selección Española Euro 2024",
		},
		{
			title:       "Camiseta Manchester City Equipación Local 2024/25",
			handle:      "camiseta-manchester-city-local-2024-25",
			bodyHTML:    "<p>La equipación local del Manchester City en su característico azul cielo de Puma para la temporada 2024/25.</p>",
			vendor:      "Puma",
			productType: "Camiseta",
			tags:        "manchester-city, local, premier-league, puma, 2024-25",
			price:       "89.95",
			compareAt:   "",
			imageSrc:    "https://cdn.football-store.example/products/manchester-city-local-2024-25/1.jpg",
			imageAlt:    "Camiseta Manchester City Local 2024/25",
		},
		{
			title:       "Camiseta Liverpool FC Equipación Local 2024/25",
			handle:      "camiseta-liverpool-local-2024-25",
			bodyHTML:    "<p>La camiseta roja del Liverpool FC para la temporada 2024/25, fabricada por Nike con tejido de alta tecnología.</p>",
			vendor:      "Nike",
			productType: "Camiseta",
			tags:        "liverpool, local, premier-league, nike, 2024-25",
			price:       "89.95",
			compareAt:   "",
			imageSrc:    "https://cdn.football-store.example/products/liverpool-local-2024-25/1.jpg",
			imageAlt:    "Camiseta Liverpool FC Local 2024/25",
		},
	}

	var productIDs []int64

	for _, seed := range seeds {
		s.productSeq++
		pid := s.productSeq
		productIDs = append(productIDs, pid)

		imgID := pid*100 + 1
		altStr := seed.imageAlt
		img := ProductImage{
			ID:         imgID,
			ProductID:  pid,
			Position:   1,
			Src:        seed.imageSrc,
			Width:      800,
			Height:     800,
			Alt:        &altStr,
			VariantIDs: []int64{},
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		optionID := pid*10 + 1
		option := ProductOption{
			ID:        optionID,
			ProductID: pid,
			Name:      "Talla",
			Position:  1,
			Values:    sizes,
		}

		var variants []Variant
		for i, size := range sizes {
			vid := pid*1000 + int64(i+1)
			sku := fmt.Sprintf("%s-%s", seed.handle, size)
			var compareAt *string
			if seed.compareAt != "" {
				c := seed.compareAt
				compareAt = &c
			}
			variants = append(variants, Variant{
				ID:                  vid,
				ProductID:           pid,
				Title:               size,
				SKU:                 sku,
				Position:            i + 1,
				Price:               seed.price,
				CompareAtPrice:      compareAt,
				InventoryQuantity:   50,
				Option1:             size,
				Weight:              0.3,
				WeightUnit:          "kg",
				RequiresShipping:    true,
				Taxable:             true,
				FulfillmentService:  "manual",
				InventoryManagement: "shopify",
				InventoryPolicy:     "deny",
				CreatedAt:           now,
				UpdatedAt:           now,
			})
		}

		p := &Product{
			ID:          pid,
			Title:       seed.title,
			Handle:      seed.handle,
			BodyHTML:    seed.bodyHTML,
			Vendor:      seed.vendor,
			ProductType: seed.productType,
			Tags:        seed.tags,
			Status:      ProductStatusActive,
			Options:     []ProductOption{option},
			Variants:    variants,
			Images:      []ProductImage{img},
			Image:       &img,
			CreatedAt:   now,
			UpdatedAt:   now,
			PublishedAt: &pub,
		}
		s.products[pid] = p
	}

	// Collections
	type collectionSeed struct {
		title      string
		handle     string
		bodyHTML   string
		productIdx []int // 0-based indices into productIDs
	}

	collections := []collectionSeed{
		{
			title:      "La Liga",
			handle:     "la-liga",
			bodyHTML:   "<p>Camisetas oficiales de los mejores equipos de La Liga española.</p>",
			productIdx: []int{0, 1, 2, 3, 4}, // Real Madrid x2, Barça x2, Atlético
		},
		{
			title:      "Premier League",
			handle:     "premier-league",
			bodyHTML:   "<p>Camisetas oficiales de los clubes de la Premier League inglesa.</p>",
			productIdx: []int{6, 7}, // Man City, Liverpool
		},
		{
			title:      "Selecciones Nacionales",
			handle:     "selecciones-nacionales",
			bodyHTML:   "<p>Equipaciones oficiales de las selecciones nacionales de fútbol.</p>",
			productIdx: []int{5}, // España
		},
		{
			title:      "Equipaciones Locales",
			handle:     "equipaciones-locales",
			bodyHTML:   "<p>Todas las camisetas de equipación local de la temporada 2024/25.</p>",
			productIdx: []int{0, 2, 4, 5, 6, 7},
		},
		{
			title:      "Equipaciones Visitante",
			handle:     "equipaciones-visitante",
			bodyHTML:   "<p>Todas las camisetas de equipación visitante de la temporada 2024/25.</p>",
			productIdx: []int{1, 3},
		},
		{
			title:      "Adidas",
			handle:     "adidas",
			bodyHTML:   "<p>Camisetas oficiales fabricadas por Adidas.</p>",
			productIdx: []int{0, 1, 5},
		},
		{
			title:      "Nike",
			handle:     "nike",
			bodyHTML:   "<p>Camisetas oficiales fabricadas por Nike.</p>",
			productIdx: []int{2, 3, 4, 7},
		},
		{
			title:      "Puma",
			handle:     "puma",
			bodyHTML:   "<p>Camisetas oficiales fabricadas por Puma.</p>",
			productIdx: []int{6},
		},
	}

	for _, cs := range collections {
		s.collectionSeq++
		cid := s.collectionSeq
		var pids []int64
		for _, idx := range cs.productIdx {
			pids = append(pids, productIDs[idx])
		}
		s.collectionProducts[cid] = pids
		s.collections[cid] = &Collection{
			ID:          cid,
			Title:       cs.title,
			Handle:      cs.handle,
			BodyHTML:    cs.bodyHTML,
			ProductIDs:  pids,
			CreatedAt:   now,
			UpdatedAt:   now,
			PublishedAt: &pub,
		}
	}
}
