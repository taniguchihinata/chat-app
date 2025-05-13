// このファイルでユーザーリストとチャットページを同時に表示している
import { useParams } from 'react-router-dom';
import UserList from './UserList';
import Chat from './Chat';
import './ChatPage.css'; // 任意のスタイルファイル



function ChatPage() {
  const { roomId } = useParams();

  return (
    <div className="chat-page">
      <div className="user-list">
        <UserList />
      </div>
      <div className="chat-area">
        <Chat roomId={roomId} />
      </div>
    </div>
  );
}

export default ChatPage;
