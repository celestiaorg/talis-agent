# Tech Stack Document

## Introduction

The Talis-API project is a lightweight, Go-based agent designed to run on Linux servers. Its primary purpose is to monitor vital system metrics such as CPU, memory, disk activity, I/O, and network operations. By exposing clearly defined HTTP endpoints, the service provides real-time data for Prometheus while also offering additional endpoints for system health, public IP detection, payload storage, and command execution. This document outlines the technology choices made to meet these requirements while ensuring reliability, ease of installation, and straightforward integration for both internal DevOps teams and open-source users.

## Frontend Technologies

Although Talis-API does not offer a traditional graphical user interface, its HTTP endpoints act as a functional user interface for interacting with the service. All responses, including those from endpoints like "/alive" and "/ip", are provided in straightforward JSON format. This approach ensures clarity and ease of use in accessing system status and metrics. The design philosophy here focuses on simplicity; by using standard HTTP communication and JSON formatting, any user with basic knowledge of web requests can easily interact with the service using tools such as cURL or any web browser, without the need for a dedicated frontend framework.

## Backend Technologies

The core of the Talis-API project is built using the Go programming language to ensure high performance and efficient resource utilization on Linux systems. The service employs the GoFiber web framework to handle HTTP routing and request processing, offering a quick and intuitive way to define endpoints such as "/metrics", "/alive", "/ip", "/payload", and "/commands". System metrics are collected using the Prometheus client library for Go, which formats the data in a manner that is readily consumable by Prometheus. Additionally, the agent reads its configuration from a YAML file, using industry-standard libraries to parse and validate settings such as the HTTP port and desired logging level. Together, these backend technologies offer a powerful combination that meets the requirements for reliable data collection, error logging, and efficient system monitoring.

## Infrastructure and Deployment

Talis-API is designed to run on Linux servers and is distributed as a .deb package, simplifying installation and integration into existing server environments. Infrastructure decisions include the use of a Makefile and custom packaging scripts to automate the build process. A GitHub Actions workflow is set up to generate the .deb package automatically with every push to the repository, ensuring continuous integration and streamlined deployment. Using Git and frequent pushes to GitHub ensures a disciplined development process with branch-based contributions, maintaining overall code quality and facilitating collaborative development.

## Third-Party Integrations

Integration with third-party tools is a key aspect of the Talis-API design. The Prometheus client is integrated to efficiently expose key system metrics to monitoring systems. Additionally, the logging functionality leverages the github.com/gofiber/fiber/v2/log package to capture important events and errors. For development and code assistance, advanced tools such as Claude 3.7 Sonnet, Claude 3.5 Sonnet, GPT o1, GPT 4o, and Cursor are integrated into the workflow to provide intelligent code generation and real-time coding suggestions. These integrations not only streamline development but also ensure that the service remains robust and maintainable over time.

## Security and Performance Considerations

Security considerations in Talis-API are centered around the need for simplicity and transparency. While endpoints like "/commands" do not currently employ authentication or access control, it is clearly documented that these endpoints should be used responsibly in trusted environments. The logging system, configured to operate at the Info level by default (unless specified otherwise in config.yaml), captures detailed error reports including network issues and misconfigurations, which aids in troubleshooting. Performance optimizations are achieved through Go’s efficient concurrency and the non-blocking nature of the GoFiber framework. As the service is primarily designed for internal monitoring purposes, the focus remains on accurate, real-time data transmission with efficient error handling and resource management.

## Conclusion and Overall Tech Stack Summary

In summary, Talis-API leverages a carefully selected technology stack that includes Go for a high-performance backend, GoFiber for easy HTTP routing, and the Prometheus client library for robust metrics collection. The configuration is managed through YAML, ensuring that settings are both human-readable and easy to modify. Infrastructure and deployment are streamlined using a Makefile, deb packaging scripts, and GitHub Actions for continuous integration. While no traditional frontend framework is used, the service’s JSON-based HTTP interface ensures that interactions remain simple, effective, and user-friendly. This blend of technologies meets the project’s goals of providing reliable, real-time monitoring with clear error logging and an uncomplicated installation process, setting Talis-API apart as a practical tool for both internal and external use.
