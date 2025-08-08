import { create } from 'zustand';
import { toast } from 'react-hot-toast';
import { fetchAgents as apiFetchAgents, fetchMessagesForAgent as apiFetchMessagesForAgent } from '../api/client';

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
            const agents = await apiFetchAgents();
            set({ agents, isLoading: false });
        } catch (error) {
            set({ error: 'Failed to fetch agents', isLoading: false });
            toast.error('获取Agent列表失败！');
        }
    },
    fetchMessagesForAgent: async (agentId) => {
        set({ isLoading: true, error: null });
        try {
            const messages = await apiFetchMessagesForAgent(agentId);
            set({ messages, isLoading: false });
        } catch (error) {
            set({ error: 'Failed to fetch messages', isLoading: false });
            toast.error('获取消息失败！');
        }
    },
}));
