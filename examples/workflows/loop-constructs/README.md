# Loop Constructs Example

This example demonstrates all the loop constructs available in go-agent-kit for repetitive workflow execution.

## What it does

1. **Count Loop** - Executes an action a fixed number of times (batch processing)
2. **While Loop** - Continues execution while a condition remains true (accumulating points)
3. **Until Loop** - Continues execution until a condition becomes true (retry until success)
4. **Iterator Loop (Slice)** - Processes each item in a slice (file processing)
5. **Iterator Loop (Map)** - Processes each key-value pair in a map (configuration)
6. **Complex Example** - Shows nested processing with sequential flows

## Key concepts

- **NewLoop(name, count)** - Basic counted loop
- **NewLoopWhile(name, condition)** - Conditional loop with while predicate
- **NewLoopUntil(name, condition)** - Conditional loop with until predicate  
- **NewLoopOver(name, items)** - Iterator loop over collections
- **Context Keys** - Auto-set loop context:
  - `current_item` - Current item being processed
  - `current_index` - Current index (0-based for slices, key for maps)
  - `loop_iteration` - Current iteration number (1-based)
- **Composability** - Loops can contain other workflows and actions

## Running the example

```bash
go run examples/workflows/loop-constructs/main.go
```

## Sample output

```
=== Loop Constructs Example ===

--- Example 1: Count Loop ---
Processing 3 items in a batch...
  Processing item 1...
  Processing item 2...
  Processing item 3...

--- Example 2: While Loop ---
Processing until we reach 100 points...
  Iteration 1: Earned 25 points (total: 25)
  Iteration 2: Earned 25 points (total: 50)
  Iteration 3: Earned 25 points (total: 75)
  Iteration 4: Earned 25 points (total: 100)

--- Example 3: Until Loop ---
Attempting to connect until successful...
  Attempt 1: Connection failed
  Attempt 2: Connection failed
  Attempt 3: Connection successful!

--- Example 4: Loop Over Slice ---
Processing each file in the list...
  Iteration 1: Processing file[0] = config.yaml
  Iteration 2: Processing file[1] = data.json
  Iteration 3: Processing file[2] = report.pdf
  Iteration 4: Processing file[3] = image.png

--- Example 5: Loop Over Map ---
Processing configuration settings...
  Iteration 1: Setting port = 8080
  Iteration 2: Setting debug = true
  Iteration 3: Setting database = postgresql://localhost:5432/mydb
  Iteration 4: Setting timeout = 30

--- Example 6: Complex Loop with Sequential Flow ---
Processing multiple batches of data...
  Processing batch 0 with 2 users
    - User 1: user1
    - User 2: user2
  Processing batch 1 with 3 users
    - User 1: user3
    - User 2: user4
    - User 3: user5
  Processing batch 2 with 1 users
    - User 1: user6
```

## Use cases

- **Batch processing** - Process items in fixed-size batches
- **Data collection** - Accumulate data until threshold reached
- **Retry logic** - Attempt operations until success
- **File processing** - Process each file in a directory
- **Configuration** - Apply settings from key-value pairs
- **Multi-stage processing** - Complex workflows with iteration