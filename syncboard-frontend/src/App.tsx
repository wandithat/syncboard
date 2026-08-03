import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuthStore } from './store/authStore';
import { ProtectedRoute } from './components/ProtectedRoute';

// Заглушки для страниц (создадим их на следующих шагах)
const Login = () => <div className="p-10">Страница Login (в разработке)</div>;
const Register = () => <div className="p-10">Страница Register (в разработке)</div>;
const BoardsList = () => <div className="p-10">Список досок (в разработке)</div>;
const BoardView = () => <div className="p-10">Просмотр доски (в разработке)</div>;

function App() {
  const init = useAuthStore((state) => state.init);

  useEffect(() => {
    init();
  }, [init]);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        
        <Route
          path="/boards"
          element={
            <ProtectedRoute>
              <BoardsList />
            </ProtectedRoute>
          }
        />
        <Route
          path="/boards/:id"
          element={
            <ProtectedRoute>
              <BoardView />
            </ProtectedRoute>
          }
        />
        
        <Route path="/" element={<Navigate to="/boards" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;