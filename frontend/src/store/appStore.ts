import { create } from 'zustand';
import axios from 'axios';
import { toast } from 'react-hot-toast';

interface AppState {
    agents: any[];
    messages: any[];
    isLoading: boolean;
    error: string | null;
    fetchAgents: () => Promise<void>;
    fetchMessagesForAgent: (agentId: number) => Promise<void>;
}

export const useAppStore = create<AppState>((set) => ({
    agents: [],
    messages: [],
    isLoading: false,
    error: null,
    fetchAgents: async () => {
        set({ isLoading: true, error: null });
        try {
            const response = await axios.get('/api/v1/agents');
            set({ agents: response.data, isLoading: false });
        } catch (error) {
            set({ error: 'Failed to fetch agents', isLoading: false });
            toast.error('获取Agent列表失败！');
        }
    },
    fetchMessagesForAgent: async (agentId) => {
        set({ isLoading: true, error: null });
        try {
            const response = await axios.get(`/api/v1/messages/${agentId}`);
            set({ messages: response.data, isLoading: false });
        } catch (error) {
            set({ error: 'Failed to fetch messages', isLoading: false });
            toast.error('获取消息失败！');
        }
    },
}));
