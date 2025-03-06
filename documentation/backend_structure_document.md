# Backend Structure Document

## Introduction

The backend of the Talis-API service is the engine that makes the entire system tick. This service, written in Go, runs on Linux and plays a key role in monitoring system metrics such as CPU, memory, disk activity, I/O, and network usage. It provides various endpoints that not only expose these metrics for Prometheus but also allow the system to report its health status, capture public IP addresses, store arbitrary payloads, and even execute bash commands. The backend is designed to be straightforward for DevOps teams and open-source users, ensuring that both real-time data collection and system integrity remain top priorities.

## Backend Architecture

The architecture of Talis-API is built using the Go programming language with GoFiber as the web framework. This design was chosen to maximize efficiency and easy deployment on Linux servers. The structure follows a modular design pattern, where each endpoint is clearly separated in the code, allowing for straightforward testing, maintenance, and future enhancements. The service reads its configuration from a YAML file, ensuring that critical settings like the HTTP port and logging level are easy to adjust. Moreover, the incorporation of Prometheus for metrics collection and GitHub Actions for the deb packaging process highlights a strong focus on automation and continuous integration, ensuring the system remains scalable and maintainable over time.

## Database Management

While Talis-API does not rely on a traditional database system such as SQL or NoSQL, it efficiently handles data storage using the local file system. For example, the payload received via the `/payload` endpoint is written directly to disk in a predefined directory (`/etc/talis-agent/payload`). Additionally, configuration details are read from the `config.yaml` file located in `/etc/talis-agent/`, which acts as the central point of configuration management. This approach keeps the data handling simple and reliable, especially given the focused scope of the application.

## API Design and Endpoints

The API design is centered around simplicity and effectiveness. Each endpoint has been defined carefully to serve its specific purpose. The `/alive` endpoint provides a basic health check by returning a JSON response with a 200 OK status, confirming that the service is running. The `/metrics` endpoint is dedicated to exposing all the essential system performance metrics in a format compatible with Prometheus. The `/ip` endpoint is responsible for identifying and returning the public IP address or addresses of the host, returning an empty JSON object if no public IPs are found. For ingesting data, the `/payload` endpoint accepts any posted payload and saves it directly to disk without restrictions on the data type or size. Finally, the `/commands` endpoint takes incoming bash commands and executes them on the system. Although there is no authentication or advanced safety checks implemented for this endpoint at the moment, it is built to function in trusted environments while logging all error scenarios.

## Hosting Solutions

Talis-API is designed to run on Linux servers and is distributed as a .deb package, facilitating an easy installation process. The service is hosted in a way that ensures reliability and seamless integration into existing infrastructure. By leveraging GitHub Actions to build the deb package automatically on every push, the project embraces a robust continuous integration and deployment workflow. This staged packaging process, combined with the use of configuration files, simplifies the deployment and scaling process, making it a cost-effective and reliable solution for system monitoring needs.

## Infrastructure Components

The infrastructure supporting Talis-API blends several essential components to ensure performance and resilience. At its core, the system relies on the GoFiber framework to manage HTTP routing and request processing with minimal overhead. Prometheus integration ensures that system metrics are collected in real time, helping maintain a pulse on system performance. Furthermore, essential directories and configuration files are managed at the operating system level, ensuring that persistent storage is both secure and readily accessible. Critical tasks such as logging and error handling are managed using the GoFiber logging package, which writes detailed logs to a dedicated file, making troubleshooting straightforward. The use of a Makefile and comprehensive scripts for packaging, along with GitHub Actions automation, rounds out the infrastructure by streamlining the build and deployment process.

## Security Measures

Security in Talis-API is handled with a focus on transparency and responsibility. Although several endpoints, such as `/payload` and `/commands`, are currently open and do not require authentication, this decision is a conscious trade-off in favor of simplicity and ease of integration for trusted environments. Proper error logging is implemented to keep track of any issues that arise, and the system is designed to fail gracefully if the configuration file is missing or misconfigured. The absence of advanced access control is well-documented, ensuring that users are aware of the risks involved, especially when executing bash commands directly through the API. Future iterations may consider adding authentication mechanics once the tool moves beyond its initial internal and open-source release phase.

## Monitoring and Maintenance

Ongoing monitoring and maintenance are integral to the backend’s design. The service uses the logging capabilities of the GoFiber framework (`github.com/gofiber/fiber/v2/log`) to capture a wide range of events and error scenarios. Logs are generated at an Info level by default, unless specified otherwise in the configuration, ensuring that key operations and any issues during metrics collection are recorded. In addition, the project utilizes both unit and integration tests to ensure continuous and comprehensive coverage of all functionalities, from the health check and metrics endpoints to the more sensitive command execution feature. This rigorous testing approach, combined with automated continuous integration routines provided by GitHub Actions, ensures that the backend remains stable, performant, and up-to-date with the latest improvements and fixes.

## Conclusion and Overall Backend Summary

In conclusion, the backend of Talis-API is built to be both robust and adaptable, combining the power of Go with the simplicity of GoFiber for efficient HTTP request management. The service’s architecture supports a seamless flow of data across multiple endpoints, and its file-based data management strategy keeps operational complexity to a minimum. From exposing real-time metrics to handling arbitrary payloads and executing system commands, every element of the backend has been designed with clarity and maintainability in mind. By leveraging automated deployment pipelines and comprehensive logging and testing, Talis-API not only meets the immediate needs of DevOps teams but also provides a solid foundation for future enhancements and broader community use.
