import { useCallback, useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { boardsApi } from '../api/boards';
import { invitationsApi } from '../api/invitations';
import type { Board } from '../types';
import type { BoardInvitation } from '../api/invitations';
import { useAuthStore } from '../store/authStore';

export const BoardsList = () => {
  const [boards, setBoards] = useState<Board[]>([]);
  const [invitations, setInvitations] = useState<BoardInvitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [actionError, setActionError] = useState('');
  const logout = useAuthStore((state) => state.logout);

  const loadData = useCallback(async () => {
    try {
      const [boardData, invitationData] = await Promise.all([boardsApi.list(), invitationsApi.listMine()]);
      setBoards(boardData);
      setInvitations(invitationData);
    } catch (err) {
      console.error('Ошибка загрузки досок и приглашений:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void loadData(); }, [loadData]);

  const handleCreate = async (title: string) => {
    try {
      await boardsApi.create(title);
      await loadData();
      setShowModal(false);
    } catch (err) {
      console.error('Ошибка создания доски:', err);
      setActionError('Не удалось создать доску. Попробуйте ещё раз.');
      throw err;
    }
  };

  const respondToInvitation = async (invitation: BoardInvitation, accept: boolean) => {
    setActionError('');
    try {
      if (accept) await invitationsApi.accept(invitation.id);
      else await invitationsApi.decline(invitation.id);
      await loadData();
    } catch (err) {
      console.error('Ошибка обработки приглашения:', err);
      setActionError('Не удалось обработать приглашение. Обновите страницу и попробуйте снова.');
    }
  };

  if (loading) return <div className="flex justify-center items-center h-screen text-gray-500">Загрузка...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto py-4 px-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-800">Мои доски</h1>
          <div className="space-x-3">
            <button onClick={() => setShowModal(true)} className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition">+ Новая доска</button>
            <button onClick={logout} className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition">Выйти</button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto py-8 px-4 space-y-8">
        {actionError && <p className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">{actionError}</p>}

        {invitations.length > 0 && (
          <section className="rounded-lg border border-blue-200 bg-blue-50 p-5">
            <h2 className="text-lg font-semibold text-gray-900">Входящие приглашения</h2>
            <div className="mt-3 space-y-3">
              {invitations.map((invitation) => (
                <div key={invitation.id} className="flex flex-wrap items-center justify-between gap-3 rounded-md bg-white p-4 shadow-sm">
                  <div>
                    <p className="font-medium text-gray-900">{invitation.board_title}</p>
                    <p className="text-sm text-gray-500">Приглашение от {invitation.inviter_email}</p>
                  </div>
                  <div className="flex gap-2">
                    <button onClick={() => void respondToInvitation(invitation, true)} className="rounded-md bg-blue-600 px-3 py-2 text-sm text-white hover:bg-blue-700">Принять</button>
                    <button onClick={() => void respondToInvitation(invitation, false)} className="rounded-md bg-gray-100 px-3 py-2 text-sm text-gray-700 hover:bg-gray-200">Отклонить</button>
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        <section>
          <h2 className="mb-4 text-lg font-semibold text-gray-900">Доски</h2>
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
                  <p className="text-sm text-gray-500 mt-2">{new Date(board.created_at).toLocaleDateString('ru-RU')}</p>
                </Link>
              ))}
            </div>
          )}
        </section>
      </main>

      {showModal && <CreateBoardModal onClose={() => setShowModal(false)} onCreate={handleCreate} />}
    </div>
  );
};

const CreateBoardModal = ({ onClose, onCreate }: { onClose: () => void; onCreate: (title: string) => Promise<void> }) => {
  const [title, setTitle] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!title.trim()) { setError('Введите название доски.'); return; }
    setSaving(true);
    setError('');
    try {
      await onCreate(title.trim());
    } catch {
      setError('Не удалось создать доску. Проверьте соединение и попробуйте ещё раз.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div role="dialog" aria-modal="true" aria-label="Создать доску" className="bg-white p-6 rounded-lg w-full max-w-md shadow-xl">
        <h2 className="text-xl font-bold mb-4">Создать доску</h2>
        <form onSubmit={handleSubmit}>
          <input type="text" maxLength={120} value={title} onChange={(event) => setTitle(event.target.value)} placeholder="Название доски"
            className="w-full px-3 py-2 border border-gray-300 rounded-md mb-3 focus:ring-2 focus:ring-blue-500 outline-none" autoFocus />
          {error && <p className="mb-3 text-sm text-red-600">{error}</p>}
          <div className="flex justify-end space-x-3">
            <button type="button" onClick={onClose} className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200">Отмена</button>
            <button disabled={saving} type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">{saving ? 'Создание…' : 'Создать'}</button>
          </div>
        </form>
      </div>
    </div>
  );
};
