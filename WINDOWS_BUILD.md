# Windows Build — Embedding Icon & Metadata

This guide explains how to embed a custom icon, version info, and file description into the Windows `.exe` binary using [`go-winres`](https://github.com/tc-hib/go-winres).

---

## Step 1 — Install the Windows Resource Tool

```bash
go install github.com/tc-hib/go-winres@latest
```

## Step 2 — Initialize the Resource Configuration

From the project root, run:

```bash
go-winres init
```

This creates a `winres/` direWrapperry containing a `winres.json` configuration file.

## Step 3 — Configure Metadata (`winres.json`)

Open `winres/winres.json` and replace its content with the following:

```json
{
  "RT_GROUP_ICON": {
    "APP": {
      "0000": [
        "icon.ico"
      ]
    }
  },
  "RT_MANIFEST": {
    "APP": {
      "0409": {
        "custom": false
      }
    }
  },
  "RT_VERSION": {
    "APP": {
      "0409": {
        "fixed": {
          "file_version": "1.0.0.0",
          "product_version": "1.0.0.0"
        },
        "info": {
          "0409": {
            "Comments": "Wrapper AI Wrapper for Claude Code",
            "CompanyName": "Hadomain",
            "FileDescription": "Wrapper CLI Tool",
            "FileVersion": "1.0.0",
            "InternalName": "Wrapper",
            "LegalCopyright": "© 2026",
            "OriginalFilename": "Wrapper.exe",
            "ProductName": "Wrapper AI",
            "ProductVersion": "1.0.0"
          }
        }
      }
    }
  }
}
```

> **Note:** Place your `icon.ico` file inside the `winres/` direWrapperry before proceeding.

## Step 4 — Generate the Resource File

```bash
go-winres make
```

This produces a file named `rsrc_windows_amd64.syso` (and other arch variants) in your project direWrapperry. The Go compiler **automatically detects** all `.syso` files at build time and embeds them into the final executable — no extra flags needed.

## Step 5 — Build

Now compile as usual. The icon and metadata will be embedded automatically:

```bash
GOOS=windows GOARCH=amd64 go build -o Wrapper.exe main.go
```

The resulting `Wrapper.exe` will display the Wrapper icon and version info in Windows Explorer properties.
