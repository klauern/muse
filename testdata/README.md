# Test Data for Git Diff Testing

This directory contains golden files for testing git diffs with expected outputs.

## Structure

- `diffs/`: Contains sample git diff files
- `expected/`: Contains expected output files corresponding to the diffs

## Usage

When writing tests, use the files in these directories to verify that your commit message generation logic produces the expected output for given input diffs.
