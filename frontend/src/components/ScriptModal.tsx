import React, { useState, KeyboardEvent, useRef, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  TextField,
  IconButton,
  Box,
  Tooltip,
  Typography,
  Paper,
} from '@mui/material';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import CloseIcon from '@mui/icons-material/Close';
import TerminalEmulator, { TerminalEmulatorRef } from './TerminalEmulator';

interface ScriptModalProps {
  open: boolean;
  onClose: () => void;
  onScriptExecute?: (script: string) => void;
}

interface Message {
  type: 'user' | 'assistant';
  content: string;
  script?: string;
}

const ScriptModal: React.FC<ScriptModalProps> = ({ open, onClose, onScriptExecute }) => {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setTimeout(() => {
        inputRef.current?.focus();
      }, 100);
    }
  }, [open]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSubmit = async () => {
    if (!input.trim()) return;
    
    setLoading(true);
    const userMessage = input;
    setInput('');
    setMessages(prev => [...prev, { type: 'user', content: userMessage }]);

    try {
      const response = await fetch('http://localhost:8080/generate-script', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ prompt: userMessage }),
      });
      
      const data = await response.json();
      const scriptMatch = data.script.match(/```bash\n([\s\S]*?)```/);
      const cleanScript = scriptMatch ? scriptMatch[1].trim() : data.script;
      
      setMessages(prev => [...prev, { 
        type: 'assistant', 
        content: 'Here\'s your script:', 
        script: cleanScript 
      }]);
    } catch (error) {
      setMessages(prev => [...prev, { 
        type: 'assistant', 
        content: 'Sorry, I encountered an error generating the script.' 
      }]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSubmit();
    } else if (event.key === 'Escape') {
      onClose();
    }
  };

  const handleCopyToClipboard = (script: string) => {
    navigator.clipboard.writeText(script);
    setCopied(script);
    setTimeout(() => setCopied(null), 2000);
  };

  const handleExecuteScript = (script: string) => {
    onScriptExecute?.(script);
    onClose();
  };

  return (
    <Dialog 
      open={open} 
      onClose={onClose} 
      maxWidth="md" 
      fullWidth
      PaperProps={{
        sx: { 
          height: '70vh',
          display: 'flex',
          flexDirection: 'column',
          bgcolor: '#1e1e1e',
          color: '#fff',
          '& .MuiDialogTitle-root': {
            padding: '16px 24px',
          },
          '& .MuiDialogContent-root': {
            padding: 0,
          }
        }
      }}
    >
      <DialogTitle 
        sx={{ 
          borderBottom: '1px solid #333',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          bgcolor: '#252525'
        }}
      >
        <Typography variant="h6" component="h2" sx={{ fontWeight: 500 }}>
          Script Generator
        </Typography>
        <IconButton 
          onClick={onClose}
          size="small"
          sx={{ 
            color: '#fff',
            '&:hover': {
              bgcolor: 'rgba(255, 255, 255, 0.1)'
            }
          }}
        >
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent 
        sx={{ 
          display: 'flex', 
          flexDirection: 'column',
          overflow: 'hidden'
        }}
      >
        <Box 
          sx={{ 
            flex: 1, 
            overflowY: 'auto', 
            p: 3,
            display: 'flex',
            flexDirection: 'column',
            gap: 2
          }}
        >
          {messages.map((message, index) => (
            <Box 
              key={index}
              sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: message.type === 'user' ? 'flex-end' : 'flex-start',
                gap: 1
              }}
            >
              <Paper
                sx={{
                  p: 2,
                  maxWidth: '80%',
                  bgcolor: message.type === 'user' ? '#2b5c8c' : '#333',
                  borderRadius: 2
                }}
              >
                <Typography>{message.content}</Typography>
                {message.script && (
                  <Box sx={{ mt: 2 }}>
                    <Box 
                      sx={{ 
                        display: 'flex', 
                        justifyContent: 'flex-end', 
                        gap: 1, 
                        mb: 1 
                      }}
                    >
                      <Tooltip title={copied === message.script ? "Copied!" : "Copy to clipboard"}>
                        <IconButton 
                          size="small" 
                          onClick={() => handleCopyToClipboard(message.script!)}
                          sx={{ color: '#fff' }}
                        >
                          <ContentCopyIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Execute in terminal">
                        <IconButton 
                          size="small" 
                          onClick={() => handleExecuteScript(message.script!)}
                          sx={{ color: '#fff' }}
                        >
                          <PlayArrowIcon />
                        </IconButton>
                      </Tooltip>
                    </Box>
                    <Typography 
                      component="pre" 
                      sx={{ 
                        p: 2,
                        bgcolor: '#000',
                        color: '#00ff00',
                        borderRadius: 1,
                        overflow: 'auto',
                        fontSize: '0.9em',
                        fontFamily: 'monospace',
                        border: '1px solid #333',
                        '&:hover': {
                          bgcolor: '#0a0a0a'
                        }
                      }}
                    >
                      {message.script}
                    </Typography>
                  </Box>
                )}
              </Paper>
            </Box>
          ))}
          <div ref={messagesEndRef} />
        </Box>
        <Box sx={{ p: 2, borderTop: '1px solid #333' }}>
          <TextField
            fullWidth
            placeholder="Describe the script you want to generate... (Press Enter to submit)"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyPress}
            inputRef={inputRef}
            variant="outlined"
            disabled={loading}
            sx={{
              '& .MuiOutlinedInput-root': {
                color: '#fff',
                bgcolor: '#333',
                '& fieldset': {
                  borderColor: '#555',
                },
                '&:hover fieldset': {
                  borderColor: '#666',
                },
                '&.Mui-focused fieldset': {
                  borderColor: '#2b5c8c',
                },
              },
            }}
          />
        </Box>
      </DialogContent>
    </Dialog>
  );
};

export default ScriptModal; 