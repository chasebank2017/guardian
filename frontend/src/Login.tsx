import React, { useState } from 'react';
import { Box, Button, TextField, Typography, Paper } from '@mui/material';
import axios from 'axios';

export default function Login({ onLogin }: { onLogin: () => void }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      const resp = await axios.post('/login', { username, password });
      localStorage.setItem('token', resp.data.token);
      onLogin();
    } catch (err: any) {
      setError(err?.response?.data || '登录失败');
    }
  };

  return (
    <Box sx={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', bgcolor: '#f5f5f5' }}>
      <Paper sx={{ p: 4, minWidth: 320 }}>
        <Typography variant="h5" gutterBottom>系统登录</Typography>
        <form onSubmit={handleSubmit}>
          <TextField label="用户名" fullWidth margin="normal" value={username} onChange={e => setUsername(e.target.value)} />
          <TextField label="密码" type="password" fullWidth margin="normal" value={password} onChange={e => setPassword(e.target.value)} />
          {error && <Typography color="error" variant="body2">{error}</Typography>}
          <Button type="submit" variant="contained" color="primary" fullWidth sx={{ mt: 2 }}>登录</Button>
        </form>
      </Paper>
    </Box>
  );
}
