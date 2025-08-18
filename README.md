# Go Concurrent File Hasher

A high-performance, concurrent command-line tool written in Go to calculate and verify the SHA-256 hashes of files within a directory.

## Features

-   **Concurrent Hashing:** Uses Go's goroutines and channels to hash multiple files simultaneously, leveraging multi-core processors for speed.
-   **Structured Logging:** Outputs results to a clean, machine-readable JSON log file.
-   **Verification Mode:** Can verify a directory against a previously generated log file to detect changes, additions, or deletions.
-   **Customizable:** Use CLI flags to control the operation mode, target directory, log file, and number of worker threads.

## Installation

Ensure you have Go installed (version 1.21+).

1.  Clone the repository:
    `git clone <your-repo-link>`
2.  Navigate into the directory:
    `cd file-hasher`
3.  Build the executable:
    `go build`

## Usage

### Hashing a Directory

This command will scan the `my-project` directory and save the hashes to `project.log`.

```bash
./file-hasher -mode=hash -dir="./my-project" -log="project.log"