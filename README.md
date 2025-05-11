# Kafka Go Lab

Welcome to the **Kafka Go Lab**! This project is a comprehensive demonstration of using Kafka with Go, showcasing various components like producers, consumers, processors, and a Dead Letter Queue (DLQ) viewer.

## Project Structure

- **cmd/**: Contains the main applications for different services.
  - **producer/**: Produces messages to a Kafka topic.
  - **consumer/**: Consumes messages from a Kafka topic.
  - **processor/**: Processes messages and forwards them to another topic or DLQ.
  - **reporter/**: Reports on processed messages, such as calculating total tax.

- **internal/**: Contains internal packages used by the project.
  - **dlq/**: Implements the Dead Letter Queue logic for handling failed messages.

- **dlq_viewer.py**: A Python script to view and manage messages in the DLQ.

- **Dockerfile**: Defines the Docker image for the producer service.

- **docker-compose.yaml**: Configures the multi-container Docker application, including Kafka, Zookeeper, and all services.

- **script.sh**: A shell script for managing Docker containers and Kafka topics.

- **go.mod** and **go.sum**: Go module files for dependency management.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.22 or later
- Python 3.12 for the DLQ viewer

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/kafka-go-lab.git
   cd kafka-go-lab
   ```

2. **Build and start the services**:
   ```bash
   docker-compose up --build
   ```

3. **Access Kafka UI**:
   - [Kafka UI (Repanda)](http://localhost:8080)
   - [Kafka UI](http://localhost:8090)

### Usage

- **Producer**: Automatically starts and sends random order messages to the Kafka topic.
- **Consumer**: Reads messages from the Kafka topic and logs them.
- **Processor**: Processes messages, calculates tax, and sends them to another topic or DLQ if processing fails.
- **Reporter**: Summarizes the total tax calculated from processed messages.
- **DLQ Viewer**: Run the DLQ viewer to manage failed messages:
  ```bash
  docker-compose up dlq-viewer
  ```

### Retry Strategy

The project includes a manual retry strategy for failed messages using the DLQ viewer. Operators can choose to retry sending messages back to the original topic.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. 