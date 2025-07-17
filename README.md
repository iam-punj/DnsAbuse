# DnsAbuse

A lightweight, DNS-based query system that answers practical and fun queries—without loading entire web pages.  
Built in Go, optimized with caching, and accessed using simple DNS `dig` commands.

---

## 🌐 Overview

Ever searched Google just to get something simple like:

- "What's the exchange rate today?"
- "Define: minimalist"
- "Roll a dice"
- "What's my public IP?"

These searches return **megabytes of HTML, CSS, and JavaScript**—when all you really need is a few **bytes** of data.

**DNSAbuse** flips the model by delivering fast, text-based answers over DNS, making it ideal for lightweight environments, scripting, or just having fun with `dig`.

---

## ✨ Key Features

- 🔌 DNS query interface for common utilities  
- ⚡ Returns answers as TXT records (or A/AAAA where appropriate)  
- 🧠 In-memory caching mechanism for speed  
- 🔧 Configurable via YAML  
- 📦 Small and easy to deploy (written in Go)  

---

## 📦 Installation

### 1. Requirements

- [Go](https://golang.org/dl/) installed (v1.18 or newer)

### 2. Get the Project

Download the ZIP or clone:

```bash
git clone https://github.com/iam-punj/DnsAbuse.git
cd DnsAbuse
```

### 3. Build the Executable

Follow the command in `build-command.txt` or run:

```bash
go build -o dnsabuse main.go
```

This generates the `dnsabuse` binary.

### 4. Configure

Edit `config.yaml` to set:

- Listening IP and port  
- Cache settings  
- API keys (if needed)

### 5. Run the Server

```bash
./dnsabuse
```

---

## ✅ Usage

Once running, query your local DNSAbuse server using `dig`.

### 🔍 Help Query

```bash
dig @127.0.0.1 -p 5354 help TXT
```

You'll get an answer like:

```text
"generate random numbers"       "dig @127.0.0.1 -p 5354 1-100.rand"
"return digits of Pi"           "dig @127.0.0.1 -p 5354 pi"
"roll dice"                     "dig @127.0.0.1 -p 5354 1d6.dice"
"convert currency"              "dig @127.0.0.1 -p 5354 99USD-INR.fx"
"dictionary definitions"        "dig @127.0.0.1 -p 5354 fun.dict"
"your public IP"                "dig @127.0.0.1 -p 5354 ip"
```

---

## 🧪 Example Queries

### 🎲 Roll a Dice

```bash
dig @127.0.0.1 -p 5354 1d6.dice TXT
```

> Response: `"🎲 You rolled a 4"`

### 🔢 Random Number (1 to 100)

```bash
dig @127.0.0.1 -p 5354 1-100.rand TXT
```

> Response: `"Random number: 73"`

### 💱 Currency Conversion (e.g., 99 USD to INR)

```bash
dig @127.0.0.1 -p 5354 99USD-INR.fx TXT
```

> Response: `"99 USD = 8264.73 INR"`

### 📖 Dictionary Lookup

```bash
dig @127.0.0.1 -p 5354 minimalist.dict TXT
```

> Response: `"Minimalist: a person who favors a moderate approach or style..."`

### 🧮 Digits of Pi

```bash
dig @127.0.0.1 -p 5354 pi TXT
```

> Response: `"3.141592653589793..."`

---

## ⚙️ Project Structure

```
DnsAbuse/
├── main.go                  # Application entry point
├── config.yaml              # Config file
├── build-command.txt        # Build instructions
├── handlers/                # Request handlers
└── cache/                   # Caching layer
```

## 🙋 Author

**@iam-punj**  
Contributions, feedback, and suggestions welcome.

---

## 🧪 Quick Test

To test if the server is working, run:

```bash
dig @127.0.0.1 -p 5354 help 
```

---

## 📌 Notes

- DNSAbuse is intended for experimental and educational purposes.  
- Actual data sources may require API keys or rate limiting.  
- DNS-based querying should not replace APIs for production systems—this project is an exploration of protocol misuse (in a fun way!).  

---
