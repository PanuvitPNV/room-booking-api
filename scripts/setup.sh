#!/bin/bash

echo "🏨 Setting up Hotel Booking API with Docker..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
npm install

# Start PostgreSQL container
echo "🐳 Starting PostgreSQL container..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
until docker-compose exec postgres pg_isready -U postgres; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

# Create databases
echo "🗄️ Recreating databases..."

# Drop databases if they exist
docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS hotel_booking;"
docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS hotel_booking_test;"

# Create fresh databases
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE hotel_booking;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE hotel_booking_test;"

echo "✅ Databases recreated: hotel_booking, hotel_booking_test"


# Build the project
echo "🔨 Building project..."
npm run build

# Initialize database
echo "🔧 Initializing database..."
npm run init-db

echo "✅ Setup complete!"
echo ""
echo "🚀 To start the server:"
echo "   npm run dev"
echo ""
echo "🧪 To run tests:"
echo "   npm test"
echo ""
echo "📊 To run load tests:"
echo "   ./scripts/load-test.sh"
echo ""
echo "🐳 Docker services:"
echo "   PostgreSQL: localhost:5432"
echo "   Adminer (DB UI): http://localhost:8080"
echo ""
echo "🛑 To stop services:"
echo "   docker-compose down"