import React, { useEffect, useState, useRef } from 'react';
import { Box } from '@mui/material';
import ScriptModal from './components/ScriptModal';
import TerminalEmulator, { TerminalEmulatorRef } from './components/TerminalEmulator';
import './App.css';

function App() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const terminalRef = useRef<TerminalEmulatorRef>(null);

  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      // Check for Command + I (Mac) or Control + I (Windows/Linux)
      if ((event.metaKey || event.ctrlKey) && event.key === 'i') {
        event.preventDefault();
        setIsModalOpen(true);
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, []);

  const handleScriptExecution = (script: string) => {
    if (terminalRef.current) {
      terminalRef.current.addOutput(`Executing script:\n${script}\n`, 'output');
    }
  };

  return (
    <Box 
      className="App" 
      sx={{ 
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        bgcolor: '#000',
        overflow: 'hidden'
      }}
    >
      <Box sx={{ flex: 1, p: 0, height: '100%' }}>
        <TerminalEmulator
          ref={terminalRef}
          initialWorkingDirectory="~"
          onExecute={(command) => {
            if (terminalRef.current) {
              terminalRef.current.addOutput(`Executing: ${command}\n`, 'output');
            }
          }}
        />
      </Box>
      <ScriptModal 
        open={isModalOpen} 
        onClose={() => setIsModalOpen(false)}
        onScriptExecute={handleScriptExecution}
      />
    </Box>
  );
}

export default App;
