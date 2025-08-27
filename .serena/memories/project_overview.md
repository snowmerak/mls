# MLS Project Overview

## Purpose
MLS (name possibly stands for "Multi-Level Storage" or similar) is a Go library that provides an abstract tree data structure interface with disk-based implementation. The project defines interfaces for tree elements and tree operations, with a planned disk-based storage implementation.

## Current State
- The project is in early development stage
- Core interfaces are defined in `lib/tree/tree.go`
- Disk implementation skeleton exists in `lib/tree/disk/tree.go` but all methods are unimplemented (panic with "unimplemented")
- No tests, documentation, or example usage currently exist

## Repository Information
- Owner: snowmerak
- Repository: github.com/snowmerak/mls
- License: MIT License (2025)
- Current branch: main

## Core Components
1. **Tree Interface**: Defines basic tree operations (Insert, Find, Delete, Head)
2. **Element Interface**: Defines tree node operations with left/right children and counts
3. **Disk Implementation**: Planned persistent storage implementation (currently unimplemented)