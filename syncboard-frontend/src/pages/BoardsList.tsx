import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { boardsApi } from '../api/boards';
import type { Board } from '../types';
import { useAuthStore } from '../store/authStore';

export const BoardsList = () => {
  const [boards, setBoards] = useState<Board[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const logout = useAuthStore((state) => state.logout);

  useEffect(() => {
    loadBoards();
  }, []);

  const loadBoards = async () => {
    try {
      const data = await boardsApi.list();
      setBoards(data);
    } catch (err) {
      console.error('Ошибка загрузки досок:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (title: string) => {
    try {
      await boardsApi.create(title);
      await loadBoards();
      setShowModal(false);
    } catch (err) {
      console.error('Ошибка создания:', err);
    }
  };

  if (loading) return <div className="flex justify-center items-center h-screen text-gray-500">Загрузка...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto py-4 px-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-800">Мои доски</h1>
          <div className="space-x-3">
            <button onClick={() => setShowModal(true)} className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition">
              + Новая доска
            </button>
            <button onClick={logout} className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition">
              Выйти
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto py-8 px-4">
        {boards.length === 0 ? (
          <div className="text-center py-20 bg-white rounded-lg border border-dashed border-gray-300">
            <p className="text-gray-500 text-lg">У вас пока нет досок</p>
            <button onClick={() => setShowModal(true)} className="mt-4 text-blue-600 hover:underline">Создать первую доску</button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {boards.map((board) => (
              <Link key={board.id} to={`/boards/${board.id}`} className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 hover:shadow-md hover:border-blue-300 transition-all group">
                <h3 className="text-xl font-semibold text-gray-800 group-hover:text-blue-600">{board.title}</h3>
                <p className="text-sm text-gray-500 mt-2">
                  {new Date(board.created_at).toLocaleDateString('ru-RU')}
                </p>
              </Link>
            ))}
          </div>
        )}
      </main>

      {showModal && <CreateBoardModal onClose={() => setShowModal(false)} onCreate={handleCreate} />}
    </div>
  );
};

// Компонент модального окна (внутри того же файла для простоты)
const CreateBoardModal = ({ onClose, onCreate }: { onClose: () => void; onCreate: (title: string) => void }) => {
  const [title, setTitle] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (title.trim()) onCreate(title.trim());
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white p-6 rounded-lg w-full max-w-md shadow-xl">
        <h2 className="text-xl font-bold mb-4">Создать доску</h2>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Название доски"
            className="w-full px-3 py-2 border border-gray-300 rounded-md mb-4 focus:ring-2 focus:ring-blue-500 outline-none"
            autoFocus
          />
          <div className="flex justify-end space-x-3">
            <button type="button" onClick={onClose} className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200">Отмена</button>
            <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">Создать</button>
          </div>
        </form>
      </div>
    </div>
  );
};
