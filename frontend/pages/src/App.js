import './App.css';

function App() {
  return (
    <div className="App">
      <h1>新規登録</h1>

      <div className="form-row">
        <label>ユーザー名</label>
        <input type="text" />
      </div>
      
      <div className="form-row">
        <label>パスワード</label>
        <input type="text"/>
      </div>
      
      <p></p>
      <button>登録</button>
    </div>
  );
}

export default App;
