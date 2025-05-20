import React, { useState, useRef, useEffect, forwardRef, useImperativeHandle } from 'react';
import { Box, Paper, Typography } from '@mui/material';

interface TerminalEmulatorProps {
  script?: string;
  onExecute?: (command: string) => void;
  initialWorkingDirectory?: string;
}

export interface TerminalEmulatorRef {
  addOutput: (output: string, type?: 'output' | 'error') => void;
  clear: () => void;
}

interface TerminalCommand {
  name: string;
  description: string;
  execute: (args: string[]) => string;
}

const TerminalEmulator = forwardRef<TerminalEmulatorRef, TerminalEmulatorProps>((props, ref) => {
  const { script, onExecute, initialWorkingDirectory = '~' } = props;
  const [lines, setLines] = useState<Array<{ content: string; type: 'input' | 'output' | 'error' }>>([
    { content: 'Welcome to Script Generator Terminal. Type "help" for available commands.', type: 'output' },
    { content: '$ ', type: 'input' }
  ]);
  const [currentInput, setCurrentInput] = useState('');
  const [history, setHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [workingDirectory, setWorkingDirectory] = useState(initialWorkingDirectory);
  const [selectedText, setSelectedText] = useState('');
  const terminalRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Built-in commands
  const builtInCommands: TerminalCommand[] = [
    {
      name: 'clear',
      description: 'Clear the terminal screen',
      execute: () => {
        setLines([{ content: '$ ', type: 'input' }]);
        return '';
      }
    },
    {
      name: 'help',
      description: 'Show available commands',
      execute: () => {
        return `Available commands:
  clear - Clear the terminal screen
  help  - Show this help message
  pwd   - Print working directory
  cd    - Change directory (simulation)
  history - Show command history`;
      }
    },
    {
      name: 'pwd',
      description: 'Print working directory',
      execute: () => workingDirectory
    },
    {
      name: 'cd',
      description: 'Change directory (simulation)',
      execute: (args) => {
        const newPath = args[0] || '~';
        setWorkingDirectory(newPath);
        return '';
      }
    },
    {
      name: 'history',
      description: 'Show command history',
      execute: () => history.reverse().join('\n')
    }
  ];

  useImperativeHandle(ref, () => ({
    addOutput: (output: string, type: 'output' | 'error' = 'output') => {
      setLines(prev => [...prev, { content: output, type }]);
    },
    clear: () => {
      setLines([{ content: '$ ', type: 'input' }]);
    }
  }));

  useEffect(() => {
    if (script) {
      setCurrentInput(script);
    }
  }, [script]);

  useEffect(() => {
    // Scroll to bottom when lines change
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [lines]);

  // Load history from localStorage on mount
  useEffect(() => {
    const savedHistory = localStorage.getItem('terminalHistory');
    if (savedHistory) {
      setHistory(JSON.parse(savedHistory));
    }
  }, []);

  // Save history to localStorage when it changes
  useEffect(() => {
    localStorage.setItem('terminalHistory', JSON.stringify(history));
  }, [history]);

  const executeCommand = (command: string) => {
    const [cmd, ...args] = command.trim().split(' ');
    const builtInCommand = builtInCommands.find(c => c.name === cmd);

    if (builtInCommand) {
      const output = builtInCommand.execute(args);
      if (output) {
        setLines(prev => [...prev, { content: output, type: 'output' }]);
      }
    } else if (onExecute) {
      onExecute(command);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      const command = currentInput.trim();
      if (command) {
        // Add command to history
        setHistory(prev => [command, ...prev].slice(0, 50)); // Keep last 50 commands
        setHistoryIndex(-1);

        // Add command to terminal output
        setLines(prev => [
          ...prev,
          { content: `${workingDirectory} $ ${currentInput}`, type: 'input' }
        ]);

        // Execute command
        executeCommand(command);

        // Clear input and add new prompt
        setCurrentInput('');
        setLines(prev => [...prev, { content: '$ ', type: 'input' }]);
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (historyIndex < history.length - 1) {
        const newIndex = historyIndex + 1;
        setHistoryIndex(newIndex);
        setCurrentInput(history[newIndex]);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex > 0) {
        const newIndex = historyIndex - 1;
        setHistoryIndex(newIndex);
        setCurrentInput(history[newIndex]);
      } else if (historyIndex === 0) {
        setHistoryIndex(-1);
        setCurrentInput('');
      }
    } else if (e.key === 'Tab') {
      e.preventDefault();
      // Simple command auto-completion
      const input = currentInput.toLowerCase();
      const suggestion = builtInCommands.find(cmd => cmd.name.startsWith(input));
      if (suggestion) {
        setCurrentInput(suggestion.name);
      }
    } else if (e.ctrlKey && e.key === 'l') {
      e.preventDefault();
      executeCommand('clear');
    }
  };

  const handleCopy = () => {
    if (selectedText) {
      navigator.clipboard.writeText(selectedText);
    }
  };

  const handleMouseUp = () => {
    const selection = window.getSelection();
    if (selection) {
      setSelectedText(selection.toString());
    }
  };

  return (
    <Paper
      sx={{
        bgcolor: '#000',
        color: '#00ff00',
        p: 2,
        height: '400px',
        overflow: 'auto',
        fontFamily: 'monospace',
        fontSize: '14px',
        position: 'relative',
        cursor: 'text',
        '& ::selection': {
          backgroundColor: 'rgba(0, 255, 0, 0.3)',
          color: '#fff'
        }
      }}
      ref={terminalRef}
      onClick={() => inputRef.current?.focus()}
      onMouseUp={handleMouseUp}
      onKeyDown={(e) => {
        if (e.ctrlKey && e.key === 'c') {
          handleCopy();
        }
      }}
    >
      {lines.map((line, i) => (
        <Box 
          key={i} 
          sx={{ 
            whiteSpace: 'pre-wrap', 
            color: line.type === 'error' ? '#ff0000' : '#00ff00',
            '&:hover': {
              backgroundColor: 'rgba(255, 255, 255, 0.05)'
            }
          }}
        >
          {line.content}
        </Box>
      ))}
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <Typography component="span" sx={{ color: '#00ff00' }}>
          {workingDirectory} $&nbsp;
        </Typography>
        <input
          ref={inputRef}
          value={currentInput}
          onChange={(e) => setCurrentInput(e.target.value)}
          onKeyDown={handleKeyDown}
          style={{
            background: 'transparent',
            border: 'none',
            outline: 'none',
            color: '#00ff00',
            fontFamily: 'monospace',
            fontSize: '14px',
            width: '100%',
          }}
          autoFocus
        />
      </Box>
    </Paper>
  );
});

TerminalEmulator.displayName = 'TerminalEmulator';

export default TerminalEmulator; 