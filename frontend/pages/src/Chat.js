import { useParams } from 'react-router-dom';

function Chat() {
  const { id } = useParams(); // 👈 URLから相手のIDを取得

  return (
    <div>
      <h2>チャット画面</h2>
      <p>相手のユーザーID: {id}</p>
      {/* 後でここにメッセージ一覧＋入力欄を追加 */}
    </div>
  );
}

export default Chat;
