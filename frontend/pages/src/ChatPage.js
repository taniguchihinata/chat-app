import { useParams } from 'react-router-dom';
import UserList from './UserList';
import Chat from './Chat';
import './ChatPage.css';

function ChatPage({ username }) {
  const { roomId } = useParams();

  return (
    <div className="chat-page">
      <div className="user-list">
        <UserList username={username} />
      </div>
      <div className="chat-area">
        <Chat roomId={roomId} username={username} />
      </div>
    </div>
  );
}

export default ChatPage;
// このファイルでユーザーリストとチャットページを同時に表示している
