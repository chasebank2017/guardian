import React, { useEffect, useState } from 'react';
import { Box, List, ListItem, ListItemButton, ListItemText, Typography, Paper } from '@mui/material';
import { getAgents, getMessages } from './api/guardianApi';
import Login from './Login';
import './api/axiosConfig';

function App() {
  const [authed, setAuthed] = useState(!!localStorage.getItem('token'));
  const [agents, setAgents] = useState([]);
  const [selectedAgent, setSelectedAgent] = useState(null);
  const [messages, setMessages] = useState([]);

  useEffect(() => {
    if (authed) getAgents().then(setAgents);
  }, [authed]);

  useEffect(() => {
    if (selectedAgent) {
      getMessages(selectedAgent.id).then(setMessages);
    } else {
      setMessages([]);
    }
  }, [selectedAgent]);

  if (!authed) return <Login onLogin={() => setAuthed(true)} />;

  return (
    <Box sx={{ display: 'flex', height: '100vh', bgcolor: '#f5f5f5' }}>
      <Paper sx={{ width: 240, minWidth: 200, height: '100%', overflow: 'auto' }}>
        <Typography variant="h6" sx={{ p: 2 }}>Agent 列表</Typography>
        <List>
          {agents.map(agent => (
            <ListItem key={agent.id} disablePadding>
              <ListItemButton selected={selectedAgent?.id === agent.id} onClick={() => setSelectedAgent(agent)}>
                <ListItemText primary={agent.name} />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </Paper>
      <Box sx={{ flex: 1, p: 3 }}>
        <Typography variant="h6">聊天记录</Typography>
        <Paper sx={{ mt: 2, p: 2, minHeight: 400 }}>
          {messages.length === 0 ? (
            <Typography color="text.secondary">请选择左侧Agent查看消息</Typography>
          ) : (
            messages.map((msg, idx) => (
              <Box key={idx} sx={{ mb: 2 }}>
                <Typography variant="body2" color="text.secondary">
                  {new Date(msg.timestamp).toLocaleString()}
                </Typography>
                <Typography variant="body1">{msg.content}</Typography>
              </Box>
            ))
          )}
        </Paper>
      </Box>
    </Box>
  );
}

export default App;
