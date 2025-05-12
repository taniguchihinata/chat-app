//import './App.css';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import Signup from './Signup';
import Login from './Login';
import UserList from './UserList';
import ChatPage from './ChatPage';  

function App() {
  const location = useLocation();
  const showNav = location.pathname === '/' || location.pathname === '/signup';

  return (
    <div className="App">
    {showNav && (
      <nav>
        <Link to="/signup"><button>サインアップへ</button></Link>
        <Link to="/"><button>ログインへ</button></Link>
      </nav>
      )}

      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/users" element={<UserList />} />
        <Route path="/chat/:username" element={<ChatPage />} />
      </Routes>
    </div>
  );
}

export default App;
