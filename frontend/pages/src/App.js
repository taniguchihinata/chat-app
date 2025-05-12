import './App.css';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import Signup from './Signup';
import Login from './Login';
import UserList from './UserList';

function App() {
  const location = useLocation();
  const hideNav = location.pathname === '/users';

  return (
    <div className="App">
      {!hideNav && (
        <nav>
          <Link to="/signup"><button>サインアップへ</button></Link>
          <Link to="/"><button>ログインへ</button></Link>
        </nav>
      )}

      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/users" element={<UserList />} />
      </Routes>
    </div>
  );
}

export default App;
