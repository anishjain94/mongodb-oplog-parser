# MongoDB Oplog to SQL Parser

## Overview

This Go project implements a parser that converts MongoDB operation log (oplog) entries into equivalent SQL statements. It's designed to facilitate data migration from MongoDB to a relational database management system (RDBMS).

## Background

This tool simplifies the migration process by translating MongoDB's JSON documents into corresponding SQL operations.

## Features

- Parses MongoDB oplog entries
- Generates equivalent SQL statements for:
  - Insert operations
  - Update operations
  - Delete operations
- Focuses on core data manipulation operations, excluding metadata fields like version and timestamp

## Inspiration

This project draws inspiration from the open-source tool Stampede but provides a separate implementation in Go.

## Sample Oplog Entry

```json
{
  "op": "i",
  "ns": "test.student",
  "o": {
    "_id": "635b79e231d82a8ab1de863b",
    "name": "Selena Miller",
    "roll_no": 51,
    "is_graduated": false,
    "date_of_birth": "2000-01-30"
  }
}
```

## Getting Started

1. Clone the repository:
   ```
   git clone https://github.com/anishjain94/mongodb-oplog-parser.git
   ```
2. Navigate to the project directory:
   ```
   cd mongodb-oplog-parser
   ```
3. Install dependencies (if any):
   ```
   go mod tidy
   ```
4. Run the program:
   ```
   go run main.go
   ```
