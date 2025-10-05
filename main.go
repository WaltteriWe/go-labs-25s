package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Animal represents an animal in our database
type Animal struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	Species    string             `json:"species" bson:"species"`
	Age        int                `json:"age" bson:"age"`
	Color      string             `json:"color" bson:"color"`
	Weight     float64            `json:"weight" bson:"weight"`
	LocationID string             `json:"location_id,omitempty" bson:"location_id,omitempty"`
	Created    time.Time          `json:"created" bson:"created"`
	Updated    time.Time          `json:"updated" bson:"updated"`
}

// Location represents a location in our database
type Location struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Address     string             `json:"address" bson:"address"`
	City        string             `json:"city" bson:"city"`
	Country     string             `json:"country" bson:"country"`
	Coordinates struct {
		Lat float64 `json:"lat" bson:"lat"`
		Lng float64 `json:"lng" bson:"lng"`
	} `json:"coordinates" bson:"coordinates"`
	Created time.Time `json:"created" bson:"created"`
	Updated time.Time `json:"updated" bson:"updated"`
}

type Article struct {
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type Articles []Article
type Animals []Animal
type Locations []Location

// Global database variables
var client *mongo.Client
var database *mongo.Database
var animalsCollection *mongo.Collection
var locationsCollection *mongo.Collection

// Load environment variables from .env file
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
}

// Database connection
func connectDatabase() {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is required. Please check your .env file.")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	log.Println("Connected to MongoDB successfully!")
	database = client.Database("palvelinohjelmointi")
	animalsCollection = database.Collection("animals")
	locationsCollection = database.Collection("locations")
}

// GET /animals - Get all animals
func allAnimals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := animalsCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch animals", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var animals Animals
	if err = cursor.All(ctx, &animals); err != nil {
		http.Error(w, "Failed to decode animals", http.StatusInternalServerError)
		return
	}

	if animals == nil {
		animals = Animals{}
	}

	fmt.Println("Endpoint Hit: allAnimals")
	json.NewEncoder(w).Encode(animals)
}

// GET /animals/{id} - Get single animal
func getAnimal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/animals/")
	if id == "" {
		http.Error(w, "Animal ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid animal ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var animal Animal
	err = animalsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&animal)
	if err != nil {
		http.Error(w, "Animal not found", http.StatusNotFound)
		return
	}

	fmt.Println("Endpoint Hit: getAnimal")
	json.NewEncoder(w).Encode(animal)
}

// POST /animals - Create new animal
func createAnimal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var animal Animal
	if err := json.NewDecoder(r.Body).Decode(&animal); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if animal.Name == "" || animal.Species == "" {
		http.Error(w, "Animal name and species are required", http.StatusBadRequest)
		return
	}

	animal.ID = primitive.NewObjectID()
	animal.Created = time.Now()
	animal.Updated = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := animalsCollection.InsertOne(ctx, animal)
	if err != nil {
		http.Error(w, "Failed to create animal", http.StatusInternalServerError)
		return
	}

	animal.ID = result.InsertedID.(primitive.ObjectID)

	fmt.Println("Endpoint Hit: createAnimal")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(animal)
}

// PUT /animals/{id} - Update animal
func updateAnimal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/animals/")
	if id == "" {
		http.Error(w, "Animal ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid animal ID format", http.StatusBadRequest)
		return
	}

	var updateData Animal
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"updated": time.Now(),
		},
	}

	if updateData.Name != "" {
		update["$set"].(bson.M)["name"] = updateData.Name
	}
	if updateData.Species != "" {
		update["$set"].(bson.M)["species"] = updateData.Species
	}
	if updateData.Age > 0 {
		update["$set"].(bson.M)["age"] = updateData.Age
	}
	if updateData.Color != "" {
		update["$set"].(bson.M)["color"] = updateData.Color
	}
	if updateData.Weight > 0 {
		update["$set"].(bson.M)["weight"] = updateData.Weight
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := animalsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		http.Error(w, "Failed to update animal", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Animal not found", http.StatusNotFound)
		return
	}

	var updatedAnimal Animal
	err = animalsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updatedAnimal)
	if err != nil {
		http.Error(w, "Failed to fetch updated animal", http.StatusInternalServerError)
		return
	}

	fmt.Println("Endpoint Hit: updateAnimal")
	json.NewEncoder(w).Encode(updatedAnimal)
}

// DELETE /animals/{id} - Delete animal
func deleteAnimal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/animals/")
	if id == "" {
		http.Error(w, "Animal ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid animal ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := animalsCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		http.Error(w, "Failed to delete animal", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Animal not found", http.StatusNotFound)
		return
	}

	fmt.Println("Endpoint Hit: deleteAnimal")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Animal deleted successfully",
		"id":      id,
	})
}

// GET /locations - Get all locations
func allLocations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := locationsCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch locations", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var locations Locations
	if err = cursor.All(ctx, &locations); err != nil {
		http.Error(w, "Failed to decode locations", http.StatusInternalServerError)
		return
	}

	if locations == nil {
		locations = Locations{}
	}

	fmt.Println("Endpoint Hit: allLocations")
	json.NewEncoder(w).Encode(locations)
}

// GET /locations/{id} - Get single location
func getLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/locations/")
	if id == "" {
		http.Error(w, "Location ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid location ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var location Location
	err = locationsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&location)
	if err != nil {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	fmt.Println("Endpoint Hit: getLocation")
	json.NewEncoder(w).Encode(location)
}

// POST /locations - Create new location
func createLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var location Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if location.Name == "" {
		http.Error(w, "Location name is required", http.StatusBadRequest)
		return
	}

	location.ID = primitive.NewObjectID()
	location.Created = time.Now()
	location.Updated = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := locationsCollection.InsertOne(ctx, location)
	if err != nil {
		http.Error(w, "Failed to create location", http.StatusInternalServerError)
		return
	}

	location.ID = result.InsertedID.(primitive.ObjectID)

	fmt.Println("Endpoint Hit: createLocation")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(location)
}

// PUT /locations/{id} - Update location
func updateLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/locations/")
	if id == "" {
		http.Error(w, "Location ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid location ID format", http.StatusBadRequest)
		return
	}

	var updateData Location
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"updated": time.Now(),
		},
	}

	if updateData.Name != "" {
		update["$set"].(bson.M)["name"] = updateData.Name
	}
	if updateData.Address != "" {
		update["$set"].(bson.M)["address"] = updateData.Address
	}
	if updateData.City != "" {
		update["$set"].(bson.M)["city"] = updateData.City
	}
	if updateData.Country != "" {
		update["$set"].(bson.M)["country"] = updateData.Country
	}
	if updateData.Coordinates.Lat != 0 || updateData.Coordinates.Lng != 0 {
		update["$set"].(bson.M)["coordinates"] = updateData.Coordinates
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := locationsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		http.Error(w, "Failed to update location", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	var updatedLocation Location
	err = locationsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updatedLocation)
	if err != nil {
		http.Error(w, "Failed to fetch updated location", http.StatusInternalServerError)
		return
	}

	fmt.Println("Endpoint Hit: updateLocation")
	json.NewEncoder(w).Encode(updatedLocation)
}

// DELETE /locations/{id} - Delete location
func deleteLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/locations/")
	if id == "" {
		http.Error(w, "Location ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid location ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := locationsCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		http.Error(w, "Failed to delete location", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	fmt.Println("Endpoint Hit: deleteLocation")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Location deleted successfully",
		"id":      id,
	})
}

func allArticles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Return empty articles array - all data should come from database
	articles := Articles{}

	fmt.Println("Endpoint Hit: allArticles")
	json.NewEncoder(w).Encode(articles)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	{
		"message": "Welcome to the Animals & Locations API!",
		"version": "1.0.0",
		"database": "All data comes from MongoDB",
		"endpoints": {
			"GET /animals": "Get all animals from database",
			"GET /animals/{id}": "Get animal by ID from database", 
			"POST /animals": "Create new animal in database",
			"PUT /animals/{id}": "Update animal in database",
			"DELETE /animals/{id}": "Delete animal from database",
			"GET /locations": "Get all locations from database",
			"GET /locations/{id}": "Get location by ID from database", 
			"POST /locations": "Create new location in database",
			"PUT /locations/{id}": "Update location in database",
			"DELETE /locations/{id}": "Delete location from database",
			"GET /articles": "Get articles (currently empty - add your own data)"
		}
	}`)
	fmt.Println("Endpoint Hit: homePage")
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/articles", allArticles)

	// Animals endpoints with method routing
	http.HandleFunc("/animals", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			allAnimals(w, r)
		case "POST":
			createAnimal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/animals/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getAnimal(w, r)
		case "PUT":
			updateAnimal(w, r)
		case "DELETE":
			deleteAnimal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server starting on http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func main() {
	fmt.Println("Starting Animals API Server...")
	loadEnv() // Load .env file
	connectDatabase()
	defer client.Disconnect(context.Background())
	handleRequests()
}
