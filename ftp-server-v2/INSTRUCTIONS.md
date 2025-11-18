# FTP Server Instructions

This project provides a simple FTP server and client written in Go.
It supports:

- USER / PASS authentication
- LIST (requires PASV)
- RETR (file download)
- PASV passive mode
- Directory sharing via -dir flag

## 1. Build the Program
```bash
go build -o ftpserver main.go
```

## 2. Start the FTP Server

Choose which directory you want to share:
```bash
./ftpserver -mode=server -port=:2121 -dir=/path/to/share
```

Example:
```bash
./ftpserver -mode=server -port=:2121 -dir=./shared
```

The server will listen on:
```bash
localhost:2121
```

## 3. Run the FTP Client

Start the client:
```bash
./ftpserver -mode=client -addr=localhost:2121
```

You will enter the FTP shell:
```bash
ftp>
```


## 4. Basic FTP Commands
Login (any username and password works)
```bash
USER anything
PASS anything
```


Enter passive mode

Required for LIST and RETR:
```bash
PASV
```

List files in the shared directory
```bash
LIST
```

Download a file
```bash
RETR filename.ext
```

Example:
```bash
PASV
RETR example.txt
```
Show current directory
```bash
PWD
```

Quit the session
```bash
QUIT
```


## 5. Notes

- The server shares only the directory given by -dir.
- LIST shows only the current directory (not recursive).
- Subdirectory files can still be downloaded using paths:
```bash
RETR folder/file.txt
```

- You must run PASV before LIST or RETR.
## 6. In-Client Help

You can type:
```bash
HELP
```

to see the supported commands.