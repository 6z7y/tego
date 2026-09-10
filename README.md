# tego

A multi-threaded download manager written in Go using *mmap*.

## Features

- **Multi-threaded** — downloads chunks in parallel with configurable thread count
- **mmap-based** — chunks written directly into the memory-mapped output file, no temp files, no merge step
- **Real-time progress** — live percentage
- **Zero intermediate I/O** — data goes straight from network → mmap → disk

## How It Works

1. Fetch file size via a `HEAD` request (`Content-Length`)
2. Pre-allocate the output file with `Truncate(size)`
3. `mmap` the entire file into memory (`RDWR`)
4. Split the file into fixed-size chunks (e.g. 10 MB)
5. Each worker downloads one chunk via `Range: bytes=start-end` and `copy()`s the response directly into the mmap slice at the correct offset
6. No merging — the file is complete when all chunks finish

```
                    mmap (whole file)
    ┌──────────┬──────────┬──────────┬──────────┐
    │ chunk 0  │ chunk 1  │ chunk 2  │ chunk 3  │
    └──────────┴──────────┴──────────┴──────────┘
         ▲          ▲          ▲          ▲
       thread     thread     thread     thread
      (range)    (range)    (range)    (range)
```

## get && Build


```bash
git clone https://github.com/6z7y/tego.git
make
```

## Install

```bash
sudo make install
```

## Usage

```bash
tego <URL>
```

# please read -h for understand used
