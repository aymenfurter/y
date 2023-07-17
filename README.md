# y - A OpenAI Code Analyzer

This is a Go program designed to perform static code analysis and identify potential security vulnerabilities using the OpenAI's GPT-4 model. The program works by traversing through all files in a given directory and scanning each file based on its language type (".go", ".py", ".js", ".java", etc). 

## Key Features

- Scans files in various languages (".go", ".py", ".js", ".java", etc).
- Uses the GPT-4 model from OpenAI to analyze the code.
- Provides a security score on the basis of severity (0-10).
- Highlights files with high severity scores (9/10 or 10/10).