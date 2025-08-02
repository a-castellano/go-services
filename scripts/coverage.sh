#!/bin/bash

# Coverage script for go-services project
# This script generates code coverage reports for all packages in the project
# Usage: ./scripts/coverage.sh [html]
# If 'html' argument is provided, generates HTML coverage report

# Create cover directory if it doesn't exist
mkdir -p cover

# Initialize coverage report with mode header
echo 'mode: count' > cover/coverage.report

# Get list of all packages in the project (excluding vendor directory)
PKG_LIST=$(go list ./... | grep -v /vendor/ | grep 'a-castellano')

# Run tests with coverage for each package
for package in ${PKG_LIST}; do
    # Generate coverage profile for each package
    # -covermode=count: counts how many times each statement is executed
    # -coverprofile: specifies output file for coverage data
    go test --tags=integration_tests -covermode=count -coverprofile "cover/${package##*/}.cov" "$package" ;
done

# Combine all coverage files into a single report
# tail -q -n +2: removes the mode line from each file (starting from line 2)
# >>: appends to the main coverage report
tail -q -n +2 cover/*.cov >> cover/coverage.report

# Display coverage summary in terminal
go tool cover -func=cover/coverage.report

# Generate HTML coverage report if requested
go tool cover -html=cover/coverage.report -o coverage.html
