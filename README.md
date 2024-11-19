# Doozip - File Archiving and Email Sender API

**Doozip** is a Go-based API that allows users to perform file archiving operations and send files via email. It supports the ability to zip multiple files and send them to a list of recipients via email.

## Features

1. **File Archiving**:
   - Upload a file and get its archive information.
   - Upload multiple files and compress them into a zip archive.

2. **Send File via Email**:
   - Upload a file (e.g., PDF or DOCX) and provide a list of email recipients to send the file to as an email attachment.

## Requirements

- Go 1.23
- A valid SMTP configuration (Gmail or any other SMTP provider)
- Environment variables for SMTP credentials

## API Endpoints

### 1. `/api/archive/information`

This endpoint retrieves information about a `.zip` file, such as its contents.

#### Example Request:
```bash
curl -X POST http://localhost:8080/api/archive/information \
-H "Content-Type: multipart/form-data" \
-F "file=@/path/to/your/archive.zip"
```

#### Response:
Returns details about the uploaded zip file.
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
    "filename": "my_archive.zip",
    "archive_size": 4102029.312,
    "total_size": 6836715.52,
    "total_files": 2,
    "files": [
        {
            "file_path": "photo.jpg",
            "size": 2516582.4,
            "mimetype": "image/jpeg"
        },
        {
            "file_path": "directory/document.docx",
            "size": 4320133.12,
            "mimetype": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        }
    ]
}
```

### 2. `/api/archive/files`

This endpoint allows you to upload multiple files and compress them into a zip archive.

#### Example Request:
```bash
curl -X POST http://localhost:8080/api/archive/files \
-H "Content-Type: multipart/form-data" \
-F "files[]=@/path/to/your/doc.docx" \
-F "files[]=@/path/to/your/img.jpg" \
-o output.zip
```

#### Response:
Returns a generated zip file.

### 3. `/api/mail/file`

This endpoint allows you to send a file as an email attachment to a list of recipients.

#### Example Request:
```bash
curl -X POST http://localhost:8080/api/mail/file \
-H "Content-Type: multipart/form-data" \
-F "file=@/path/to/your/file.pdf" \
-F "emails=recipient1@example.com,recipient2@example.com"
```

#### Response:
Returns a success message after sending the email to the recipients.

## Project Structure

```
.
├── cmd
│   └── doozip
│       └── main.go
├── config
│   └── config.yml
├── curl.txt
├── internal
│   ├── config
│   │   ├── config.go
│   │   └── config_test.go
│   ├── entities
│   │   └── entities.go
│   ├── handlers
│   │   ├── information.go
│   │   ├── mail.go
│   │   └── response.go
│   ├── logger
│   │   └── slog.go
│   ├── repositories
│   │   ├── archive.go
│   │   └── mail.go
│   ├── services
│   │   ├── archive.go
│   │   └── mail.go
│   └── utils
│       └── root.go
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── README.md
```

### Folder Descriptions:

- **`cmd`**: Contains the "main entry point" to the application (`main.go`), i wanted to run it with `.`, so i put `main.go` in the root folder, and it will call the cmd/main.go.
- **`config`**: Holds configuration files (`config.yml` for app settings).
- **`internal`**: Contains core application logic, such as handlers, services, repositories, and utilities.
- **`Makefile`**: Defines commands for building and running the application.
- **`curl.txt`**: Example cURL commands for testing the API endpoints.

## Getting Started

### 1. Clone the Repository

Clone the repository to your local machine:

```
git clone https://github.com/ab-dauletkhan/doozip.git
cd doozip
```

### 2. Set Up Environment Variables

You have export SMTP credentials, they should be written inside `./config/config.yml`

```bash
export SMTP_USERNAME=your-email@gmail.com
export SMTP_PASSWORD=your-email-password
```

### 3. Install Dependencies

Install the Go dependencies:

```bash
go mod tidy
```

### 4. Run the Application

Run the Go application:

```bash
make build
./doozip
# or
make run
```

The server should now be running at `http://localhost:8080`.

### 5. Test the Endpoints

Use `curl` or Postman to test the following API endpoints.

#### Test the archive information endpoint:

```bash
curl -X POST http://localhost:8080/api/archive/information \
-H "Content-Type: multipart/form-data" \
-F "file=@/path/to/your/file.zip"
```

#### Test the archive files endpoint:

```bash
curl -X POST http://localhost:8080/api/archive/files \
-H "Content-Type: multipart/form-data" \
-F "files[]=@/path/to/your/file1.docx" \
-F "files[]=@/path/to/your/file2.jpg" \
-o output.zip
```

#### Test the send email file endpoint:

```bash
curl -X POST http://localhost:8080/api/mail/file \
-H "Content-Type: multipart/form-data" \
-F "file=@/path/to/your/file.pdf" \
-F "emails=recipient1@example.com,recipient2@example.com"
```

## Video Tutorial

Watch the YouTube video tutorial for a detailed explanation of the project:
[YouTube Tutorial](https://www.youtube.com/watch?v=example)

## Deployed Version

Access the deployed version of the API at:
[Deployed Version](http://doozip.example.com)
