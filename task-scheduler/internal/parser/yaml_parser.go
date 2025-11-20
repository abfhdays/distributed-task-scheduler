name: example_pipeline
tasks:
  - name: extract
    type: bash
    command: "echo 'Extracting data...'; sleep 2"
    dependencies: []
    max_retries: 3
    
  - name: transform
    type: bash
    command: "echo 'Transforming data...'; sleep 3"
    dependencies: ["extract"]
    max_retries: 2
    
  - name: load
    type: bash
    command: "echo 'Loading data...'; sleep 1"
    dependencies: ["transform"]
    max_retries: 1