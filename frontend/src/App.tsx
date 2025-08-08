
import React, { useEffect, useState } from 'react';
import { Toaster } from 'react-hot-toast';
import VirtualizedMessageList from './components/VirtualizedMessageList';
import { Box, List, ListItem, ListItemButton, ListItemText, Typography, Paper } from '@mui/material';
import Login from './Login';
import './api/axiosConfig';
import { useAppStore } from './store/appStore';

function App() {
  const [authed, setAuthed] = useState(!!localStorage.getItem('token'));
  const [selectedAgent, setSelectedAgent] = useState<any>(null);
  const { agents, messages, isLoading, error, fetchAgents, fetchMessagesForAgent } = useAppStore();

  useEffect(() => {
    if (authed) fetchAgents();
  }, [authed, fetchAgents]);

  useEffect(() => {
    if (selectedAgent) {
      fetchMessagesForAgent(selectedAgent.id);
    }
  }, [selectedAgent, fetchMessagesForAgent]);

  if (!authed) return <Login onLogin={() => setAuthed(true)} />;

  return (
    <>
      <Toaster position="top-right" />
      <Box sx={{ display: 'flex', height: '100vh', bgcolor: '#f5f5f5' }}>
      <Paper sx={{ width: 240, minWidth: 200, height: '100%', overflow: 'auto' }}>
        <Typography variant="h6" sx={{ p: 2 }}>Agent 列表</Typography>
        {isLoading && <Typography color="text.secondary">加载中...</Typography>}
        {error && <Typography color="error">{error}</Typography>}
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
          {isLoading ? (
            <Typography color="text.secondary">加载中...</Typography>
          ) : error ? (
            <Typography color="error">{error}</Typography>
          ) : messages.length === 0 ? (
            <Typography color="text.secondary">请选择左侧Agent查看消息</Typography>
          ) : (
            <VirtualizedMessageList messages={messages} height={600} itemSize={70} />
          )}
        </Paper>
      </Box>
    </Box>
  </>
  );
}

export default App;
