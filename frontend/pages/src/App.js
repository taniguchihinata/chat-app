import { useState } from 'react';
import Signup from './Signup';
import Login from './Login';
import './App.css';

function App() {
  const [mode, setMode] = useState('login');

  return (
    <div>
      <button onClick={() => setMode('signup')}>サインアップへ</button>
      <button onClick={() => setMode('login')}>ログインへ</button>

      {mode === 'signup' ? <Signup /> : <Login />}
    </div>
  );
}

export default App;
