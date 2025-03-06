# Talis-API Project Requirements Document

## 1. Project Overview

Talis-API is a Go-based service designed to run on Linux servers to monitor key system metrics such as CPU usage, memory consumption, disk activity, I/O, and network operations. The primary purpose of the agent is to provide real-time data collection through a range of HTTP endpoints, enabling both internal use by a DevOps team and external integration as an open-source tool. By leveraging the Prometheus client library, Talis-API continuously collects and exposes system metrics in a format that Prometheus understands, making system performance monitoring straightforward and reliable.

This project is being built to streamline operations, simplify the monitoring process, and facilitate troubleshooting by enabling real-time data transmission and detailed server insights. The key objectives include reliable metric collection, robust logging of error scenarios (especially during network disruptions), ease of installation via a .deb package, and a clean, modular design. Success will be measured by the agent’s ability to handle data accurately, manage errors effectively, and integrate seamlessly into existing DevOps workflows.

## 2. In-Scope vs. Out-of-Scope

**In-Scope:**

*   Creating a Go-based agent service that runs on Linux servers.

*   Implementing a configuration loader that reads settings (HTTP port and log level) from a `config.yaml` file located in `/etc/talis-agent/`.

*   Exposing the following HTTP endpoints:

    *   **/alive:** Returns a JSON object with a 200 OK response to confirm the agent is running.
    *   **/metrics:** Exposes system metrics (CPU, memory, disk, I/O, network) in a Prometheus-compatible format.
    *   **/ip:** Returns the current public IP address(es) in a JSON object. If multiple public IPs exist, return all; if none, return an empty object.
    *   **/payload:** Accepts and writes a POSTed payload to disk in the `/etc/talis-agent/payload` directory without restrictions on type or size.
    *   **/commands:** Accepts and executes bash commands on the system without any authentication for now.

*   Logging all significant actions and errors using `github.com/gofiber/fiber/v2/log` at the Info log level (configurable via `config.yaml`).

*   Error handling: Generating errors when the `config.yaml` file is missing/invalid and logging network issues during metrics collection.

*   Building the Linux .deb package using a Makefile, a script to create necessary configuration directories/files, and GitHub Actions to automate the build process.

*   Implementing thorough unit tests and integration tests for all endpoints and error scenarios.

*   Following branch-based development with frequent pushes to the GitHub repository <https://github.com/celestiaorg/talis-api>.

**Out-of-Scope:**

*   Implementing any authentication or access control for endpoints, including `/payload` and `/commands`, at this stage.
*   Advanced command safety checks for the `/commands` endpoint.
*   Additional configuration options beyond HTTP port and log level in the `config.yaml` file.
*   Comprehensive monitoring or alert systems beyond basic logging and error reports.
*   Complex scaling or load balancing features for high-availability setups.
*   Integration with external logging infrastructures or advanced security modules beyond basic error logging.

## 3. User Flow

A typical user interacting with Talis-API will start by ensuring that the agent is installed correctly along with the necessary configuration file in `/etc/talis-agent/config.yaml`. Once the service is running, a user or monitoring system can first hit the **/alive** endpoint to verify that the agent is up and available. This simple health check returns an HTTP 200 OK status in a JSON format, confirming that the agent is active.

After confirming the service is running, the user can progressively interact with the other endpoints. For example, Prometheus can automatically scrape the **/metrics** endpoint for real-time collection of system metrics. A user or system might query **/ip** to obtain the current public IP address(es) of the host. Additionally, if there's data to be saved or commands to be executed, the user can perform a POST to **/payload** to store arbitrary payload data on disk, or send bash commands to **/commands** where they will be executed directly on the host. The flow emphasizes a straightforward, minimal-interaction scheme with no built-in access restrictions, relying on basic error logging to track issues.

## 4. Core Features

*   **Health Check Endpoint (/alive):**

    *   Responds with a JSON object and HTTP status 200 OK to confirm service availability.

*   **Metrics Collection Endpoint (/metrics):**

    *   Uses the Prometheus client to gather and expose critical system metrics (CPU, memory, disk, I/O, network) in a format that Prometheus can understand.

*   **Public IP Endpoint (/ip):**

    *   Detects the host’s public IP address(es) and returns them in a JSON object.
    *   If multiple public IPs exist, return an array; if none, return an empty JSON object.

*   **Payload Storage Endpoint (/payload):**

    *   Accepts POST requests containing an arbitrary payload.
    *   Writes the received payload data to disk under the `/etc/talis-agent/payload` directory.

*   **Command Execution Endpoint (/commands):**

    *   Receives bash commands via POST requests and pipes these commands directly to the system for execution.
    *   No authentication or safety checks applied (use responsibly in secure environments).

*   **Configuration Loader:**

    *   Reads settings from `/etc/talis-agent/config.yaml`.
    *   Supports HTTP port (default: 25550) and log level.
    *   Generates and logs an error if the configuration file is missing or contains invalid settings.

*   **Logging and Error Handling:**

    *   Uses `github.com/gofiber/fiber/v2/log` at the Info log level (unless overridden).
    *   Logs any significant operations and errors, such as metrics collection issues and configuration problems, to a dedicated log file.

*   **Packaging and Deployment Tools:**

    *   A Makefile for building the project and creating a .deb package.
    *   A script to ensure necessary directories and the configuration file are present.
    *   GitHub Actions set up to automatically build the .deb package upon every push.

*   **Testing:**

    *   Extensive unit tests covering individual functions.
    *   Integration tests to ensure seamless interactions between endpoints and the system.
    *   Testing of error scenarios such as missing configuration files or network issues.

## 5. Tech Stack & Tools

*   **Programming Language:**

    *   Golang (Go) will be used to implement the service.

*   **Web Framework:**

    *   GoFiber (github.com/gofiber/fiber/v2) to handle HTTP service routing.

*   **Metrics Collection:**

    *   Prometheus client library for Go to collect and expose system metrics.

*   **Configuration Management:**

    *   YAML configuration files parsed from `/etc/talis-agent/config.yaml`.

*   **Packaging:**

    *   Standard .deb packaging along with a Makefile and custom script for installation.

*   **CI/CD & Testing:**

    *   GitHub Actions for continuous integration and automated deb package creation.
    *   Unit and integration testing frameworks available in Go.

*   **Development Tools & IDE Integrations:**

    *   Cursor for advanced coding support.
    *   Tools like Claude 3.7 Sonnet, Claude 3.5 Sonnet, GPT o1, and GPT 4o for AI-powered code assistance and advanced code generation.

## 6. Non-Functional Requirements

*   **Performance:**

    *   The /metrics endpoint should serve data quickly to support real-time monitoring.
    *   Response times for endpoints like /alive and /ip are expected to be near-instantaneous under normal conditions.

*   **Security:**

    *   While no authentication is implemented initially, ensure all endpoints properly handle and log errors.
    *   Execution of bash commands via the /commands endpoint should be used with caution in secure environments.

*   **Usability:**

    *   The agent must provide clear, human-readable error messages when configuration issues occur.
    *   All outputs, especially JSON responses, should be simple and standardized for easy parsing.

*   **Reliability:**

    *   Error handling must include logging of network issues and configuration problems.
    *   The agent should gracefully handle missing or invalid configuration files by logging an informative error and halting execution as needed.

*   **Scalability:**

    *   While the initial focus is on single-server monitoring, the code structure should allow easy extension for future scalability or multi-instance deployments.

## 7. Constraints & Assumptions

*   The agent will run only on Linux servers.
*   It is assumed that Prometheus will be used to scrape the /metrics endpoint.
*   No authentication or advanced security controls will be applied to the endpoints at this stage.
*   The configuration file (`config.yaml`) is expected to be present in `/etc/talis-agent/`; otherwise, the service should fail gracefully with an error.
*   The logging mechanism will use `github.com/gofiber/fiber/v2/log` at an Info log level unless overridden by the config.
*   The environment has the necessary system permissions to execute bash commands and write payloads to disk.
*   The .deb packaging process assumes that required dependencies and scripts are available on the target system during installation.
*   Testing will include unit tests and integration tests, ensuring error scenarios like network issues and misconfigurations are covered.

## 8. Known Issues & Potential Pitfalls

*   **/commands Endpoint Risks:**

    *   Executing bash commands without any authentication could lead to misuse or accidental damage. It’s important to document this risk clearly for users.

*   **Configuration File Reliance:**

    *   The agent’s startup is dependent on the presence and validity of the `config.yaml` file. Missing or invalid configuration will halt the service; thorough validation and error logging are essential for troubleshooting.

*   **Error Logging Overhead:**

    *   Excessive logging during high frequency metrics collection could affect performance. Ensure logging is performed efficiently at the defined Info level unless more detail is needed.

*   **Multiple Public IP Detection:**

    *   Handling servers with multiple public IPs could be ambiguous. Return them as an array in the JSON object, and if no public IP is detected, send an appropriately empty JSON response.

*   **Packaging Dependencies:**

    *   The .deb packaging relies on a proper Makefile, script, and CI/CD setup via GitHub Actions. Any misconfiguration in these areas may lead to failed deployments or incomplete packages.

*   **Testing Coverage:**

    *   Both unit and integration tests must be comprehensive to cover edge cases, especially error scenarios. Missing tests could lead to undetected issues in production.

This document serves as the central blueprint for the Talis-API project. It outlines the project’s purpose, the in-scope functionalities, and technical details in plain language to ensure that the subsequent technical documents—such as the Tech Stack Document, Frontend Guidelines, Backend Structure, etc.—can be generated without any ambiguity.
