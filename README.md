# Script Copilot

A Copilot-like interface for generating Linux shell scripts using AI. This tool provides a convenient way to generate shell scripts through natural language descriptions.

## Features

- 🚀 Command + I (Mac) / Control + I (Windows/Linux) shortcut to open the script generator
- 💡 Natural language to shell script conversion
- 📋 Easy copy-to-clipboard functionality
- 🔒 Secure script generation with proper error handling
- 💻 Modern, responsive UI

## Prerequisites

- Node.js (v14 or higher)
- Go (v1.16 or higher)
- OpenAI API key

## Setup

1. Clone the repository:
```bash
git clone <your-repo-url>
cd script-copilot
```

2. Set up the frontend:
```bash
cd frontend
npm install
```

3. Set up the backend:
```bash
cd ../backend
go mod tidy
```

4. Set your OpenAI API key:
```bash
export OPENAI_API_KEY=your_api_key_here
```

## Running the Application

1. Start the backend server:
```bash
cd backend
go run main.go
```

2. In a new terminal, start the frontend development server:
```bash
cd frontend
npm start
```

3. Open your browser and navigate to `http://localhost:3000`

## Usage

1. Press Command + I (Mac) or Control + I (Windows/Linux) to open the script generator
2. Type your script requirements in natural language
3. Click "Generate Script" or press Enter
4. Copy the generated script using the copy button
5. Use the script in your terminal

## Security Considerations

- The generated scripts should be reviewed before execution
- The backend uses CORS protection
- API keys are managed through environment variables
- Input validation is implemented on both frontend and backend

## License

MIT 