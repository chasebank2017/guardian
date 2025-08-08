import React from 'react';
import { FixedSizeList as List, ListChildComponentProps } from 'react-window';

interface Message {
  content: string;
  timestamp: number;
}

const MessageItem = ({ content, timestamp }: Message) => (
  <div className="message-item">
    <span className="timestamp">{new Date(timestamp).toLocaleTimeString()}</span>
    <p className="content">{content}</p>
  </div>
);

interface VirtualizedMessageListProps {
  messages: Message[];
  height?: number;
  itemSize?: number;
}

const VirtualizedMessageList: React.FC<VirtualizedMessageListProps> = ({ messages, height = 600, itemSize = 70 }) => {
  const Row = ({ index, style }: ListChildComponentProps) => {
    const message = messages[index];
    return (
      <div style={style}>
        <MessageItem {...message} />
      </div>
    );
  };
  return (
    <List
      className="message-list-container"
      height={height}
      itemCount={messages.length}
      itemSize={itemSize}
      width={"100%"}
    >
      {Row}
    </List>
  );
};

export default VirtualizedMessageList;
