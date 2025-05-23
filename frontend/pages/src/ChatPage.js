import { useParams } from 'react-router-dom';
import { useState } from 'react';
import UserList from './UserList';
import Chat from './Chat';
import './ChatPage.css';

function ChatPage({ username }) {
  const { roomId } = useParams();
  const [reloadFlag, setReloadFlag] = useState(0);

  return (
    <div className="chat-page">
      <div>
        <UserList className="user-list" username={username} reloadFlag={reloadFlag}/>
      </div>
      <div className="chat-area">
        <Chat roomId={roomId} username={username} onReadReaset={() => setReloadFlag(prev => prev + 1)} />
      </div>
    </div>
  );
}

export default ChatPage;
// このファイルでユーザーリストとチャットページを同時に表示している
