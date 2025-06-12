# Shard Manager

> ⚠️ **DRAFT IMPLEMENTATION** ⚠️
> 
> This is an early draft implementation of the Shard Manager framework. The code is not yet ready for production use and may contain bugs or incomplete features. Use at your own risk.

A distributed shard management service providing efficient and reliable shard management capabilities, based on the research paper "Shard Manager: A Generic Shard Management Framework for Geo-distributed Applications" (SIGMOD '21).

## Overview

The Shard Manager is a service that helps manage and coordinate shards in distributed systems. It provides a gRPC-based API for shard management operations and uses PostgreSQL for persistent storage. This implementation is based on the research paper published in SIGMOD '21, which presents a generic framework for managing shards in geo-distributed applications.

## Research Paper

This project implements concepts from the paper:
"Shard Manager: A Generic Shard Management Framework for Geo-distributed Applications"
Authors: [Authors from the paper]
Published in: SIGMOD '21
DOI: [10.1145/3477132.3483546](https://dl.acm.org/doi/pdf/10.1145/3477132.3483546)

## Prerequisites

- Go 1.23.0 or later
- PostgreSQL
- Protocol Buffers compiler (protoc)
- Go plugins for Protocol Buffers

## Installation

1. Clone the repository:
```bash
git clone https://github.com/seaweedfs/shardmanager.git
cd shardmanager
```

2. Install dependencies:
```bash
go mod download
```

3. Generate Protocol Buffer code:
```bash
make proto
```

## Building

To build the project:
```bash
make build
```

## Testing

Run the test suite:
```bash
make test
```

## Project Structure

- `cmd/` - Command-line applications
- `server/` - Server implementation
- `shardmanagerpb/` - Generated Protocol Buffer code
- `db/` - Database-related code
- `shardmanager.proto` - Protocol Buffer definitions
- `schema.sql` - Database schema

## Development

### Protocol Buffer Generation

The project uses Protocol Buffers for API definitions. To regenerate the Protocol Buffer code:

```bash
make proto
```

### Database Setup

1. Create a PostgreSQL database
2. Run the schema.sql file to set up the database schema:
```bash
psql -d your_database_name -f schema.sql
```

## License

This project is licensed under the terms of the included LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 