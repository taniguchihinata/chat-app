import UserList from './UserList';
import Chat from './Chat';
import './ChatPage.css'; // 任意のスタイルファイル

function ChatPage() {
  return (
    <div className="chat-page">
      <div className="user-list">
        <UserList />
      </div>
      <div className="chat-area">
        <Chat />
      </div>
    </div>
  );
}

export default ChatPage;
