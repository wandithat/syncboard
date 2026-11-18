import { useCallback, useEffect, useState } from 'react';
import type { FormEvent, ReactNode } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import type { CollisionDetection, DragStartEvent, DragOverEvent, DragEndEvent } from '@dnd-kit/core';
import { SortableContext, horizontalListSortingStrategy, arrayMove } from '@dnd-kit/sortable';
import { boardsApi } from '../api/boards';
import { columnsApi } from '../api/columns';
import { cardsApi } from '../api/cards';
import { invitationsApi } from '../api/invitations';
import type { Board, Card, ColumnWithCards } from '../types';
import type { BoardMember } from '../api/invitations';
import { SortableColumn } from '../components/SortableColumn';
import { useBoardState } from '../hooks/useBoardState';
import { useAuthStore } from '../store/authStore';

interface BoardData extends Board {
  columns: ColumnWithCards[];
}

type CreationModal = { kind: 'column' } | { kind: 'card'; columnId: string } | null;

export const BoardView = () => {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(true);
  const [initialBoard, setInitialBoard] = useState<BoardData | null>(null);
  const token = useAuthStore((state) => state.token);
  const token = useAuthStore((state) => state.token);

  const loadBoard = useCallback(async () => {
    if (!id) return;
    try {
      const boardData = await boardsApi.get(id);
      const columns = await columnsApi.getByBoard(id);
      const columnsWithCards: ColumnWithCards[] = await Promise.all(
        columns.map(async (col) => ({ ...col, cards: await cardsApi.getByColumn(col.id) }))
      );
      setInitialBoard({ ...boardData, columns: columnsWithCards });
    } catch (err) {
      console.error('Ошибка загрузки доски:', err);
      setInitialBoard(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { void loadBoard(); }, [loadBoard]);

  useEffect(() => {
    if (!id || !token) return;
    const wsBase = (import.meta.env.VITE_WS_URL || 'ws://localhost:8080').replace(/\/$/, '');
    let socket: WebSocket | null = null;
    let reconnectTimer: number | undefined;
    let refreshTimer: number | undefined;
    let retryDelay = 1000;
    let stopped = false;

    const connect = () => {
      if (stopped) return;
      socket = new WebSocket(`${wsBase}/ws/boards/${id}?token=${encodeURIComponent(token)}`);
      socket.onopen = () => { retryDelay = 1000; void loadBoard(); };
      socket.onmessage = () => {
        window.clearTimeout(refreshTimer);
        refreshTimer = window.setTimeout(() => void loadBoard(), 120);
      };
      socket.onclose = () => {
        if (stopped) return;
        reconnectTimer = window.setTimeout(connect, retryDelay);
        retryDelay = Math.min(retryDelay * 2, 15000);
      };
      socket.onerror = () => socket?.close();
    };
    connect();
    return () => {
      stopped = true;
      window.clearTimeout(reconnectTimer);
      window.clearTimeout(refreshTimer);
      socket?.close();
    };
  }, [id, token, loadBoard]);

  if (loading) return <div className="flex h-screen items-center justify-center">Загрузка...</div>;
  if (!initialBoard) return <div className="flex h-screen items-center justify-center">Доска не найдена или у вас нет доступа</div>;
  return <BoardContent initialBoard={initialBoard} onReload={loadBoard} />;
};

const columnCollisionDetection: CollisionDetection = (args) => {
  if (args.active.data.current?.type === 'column') {
    return closestCorners({
      ...args,
      droppableContainers: args.droppableContainers.filter(
        (container) => container.data.current?.type === 'column'
      ),
    });
  }
  return closestCorners(args);
};

interface BoardContentProps {
  initialBoard: BoardData;
  onReload: () => void;
}

const BoardContent = ({ initialBoard, onReload }: BoardContentProps) => {
  const { board, reorderColumns, moveCardWithinColumn, moveCardBetweenColumns,
    addCardToColumn, removeCardFromColumn } = useBoardState(initialBoard);
  const [activeCard, setActiveCard] = useState<Card | null>(null);
  const [activeColumn, setActiveColumn] = useState<ColumnWithCards | null>(null);
  const [creationModal, setCreationModal] = useState<CreationModal>(null);
  const [creationTitle, setCreationTitle] = useState('');
  const [creationError, setCreationError] = useState('');
  const [saving, setSaving] = useState(false);
  const [deleteCardId, setDeleteCardId] = useState<string | null>(null);
  const [showMembers, setShowMembers] = useState(false);
  const [members, setMembers] = useState<BoardMember[]>([]);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteError, setInviteError] = useState('');
  const [inviteSuccess, setInviteSuccess] = useState('');
  const [inviting, setInviting] = useState(false);
  const canInvite = board.can_invite === true;

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));
  const sortedColumns = [...board.columns].sort((a, b) => a.position - b.position);

  const findColumnByCardId = (cardId: string) => board.columns.find((col) => col.cards.some((card) => card.id === cardId));

  const handleDragStart = (event: DragStartEvent) => {
    if (event.active.data.current?.type === 'card') {
      setActiveCard(event.active.data.current.card as Card);
    } else if (event.active.data.current?.type === 'column') {
      setActiveColumn(event.active.data.current.column as ColumnWithCards);
    }
  };

  const handleDragOver = (event: DragOverEvent) => {
    const { active, over } = event;
    if (!over || active.data.current?.type !== 'card') return;
    if (!over || active.data.current?.type !== 'card') return;
    const activeId = String(active.id);
    const overId = String(over.id);
    if (activeId === overId) return;
    const activeColumn = findColumnByCardId(activeId);
    const overColumn = findColumnByCardId(overId) || board.columns.find((column) => column.id === overId);
    if (!activeColumn || !overColumn) return;
    const overIndex = overColumn.cards.findIndex((card) => card.id === overId);
    if (activeColumn.id === overColumn.id) {
      const activeIndex = activeColumn.cards.findIndex((card) => card.id === activeId);
      if (activeIndex >= 0 && overIndex >= 0 && activeIndex !== overIndex) {
        moveCardWithinColumn(activeColumn.id, activeIndex, overIndex);
      }
    } else {
      moveCardBetweenColumns(activeId, activeColumn.id, overColumn.id,
        overIndex >= 0 ? overIndex : overColumn.cards.length);
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveCard(null);
    setActiveColumn(null);
    if (!over) { onReload(); return; }

    if (active.data.current?.type === 'column') {
      const oldIndex = sortedColumns.findIndex((column) => column.id === String(active.id));
      const newIndex = sortedColumns.findIndex((column) => column.id === String(over.id));
      if (oldIndex < 0 || newIndex < 0 || oldIndex === newIndex) return;
      const nextColumns = arrayMove(sortedColumns, oldIndex, newIndex);
      const nextIds = nextColumns.map((column) => column.id);
      reorderColumns(nextIds);
      try {
        await columnsApi.reorder(board.id, nextIds);
      } catch (err) {
        console.error('Ошибка сохранения порядка колонок:', err);
        onReload();
      }
      return;
    }

    if (active.data.current?.type !== 'card') return;
    const activeId = String(active.id);
    const column = findColumnByCardId(activeId);
    const card = column?.cards.find((item) => item.id === activeId);
    if (!column || !card) return;
    try {
      await cardsApi.move(activeId, column.id, card.position);
      await cardsApi.move(activeId, column.id, card.position);
    } catch (err) {
      console.error('Ошибка сохранения карточки:', err);
      onReload();
    }
  };

  const openCreationModal = (modal: CreationModal) => {
    setCreationModal(modal);
    setCreationTitle('');
    setCreationError('');
  };

  const submitCreation = async (event: FormEvent) => {
    event.preventDefault();
    const title = creationTitle.trim();
    if (!title) { setCreationError('Введите название.'); return; }
    setSaving(true);
    setCreationError('');
    try {
      if (creationModal?.kind === 'column') {
        await columnsApi.create(board.id, title);
        await onReload();
      } else if (creationModal?.kind === 'card') {
        const card = await cardsApi.create(creationModal.columnId, title);
        addCardToColumn(creationModal.columnId, card);
      }
      setCreationModal(null);
    } catch (err) {
      console.error('Ошибка создания:', err);
      setCreationError('Не удалось создать элемент. Попробуйте ещё раз.');
    } finally {
      setSaving(false);
    }
  };

  const confirmDeleteCard = async () => {
    if (!deleteCardId) return;
    try {
      await cardsApi.delete(deleteCardId);
      removeCardFromColumn(deleteCardId);
      setDeleteCardId(null);
    } catch (err) {
      console.error('Ошибка удаления карточки:', err);
    }
  };

  const openMembers = async () => {
    setShowMembers(true);
    setInviteError('');
    setInviteSuccess('');
    try {
      setMembers(await invitationsApi.listMembers(board.id));
    } catch (err) {
      console.error('Ошибка загрузки участников:', err);
    }
  };

  const submitInvitation = async (event: FormEvent) => {
    event.preventDefault();
    setInviting(true);
    setInviteError('');
    setInviteSuccess('');
    try {
      await invitationsApi.create(board.id, inviteEmail.trim());
      setInviteSuccess(`Приглашение для ${inviteEmail.trim()} создано. Оно появится у пользователя в списке приглашений.`);
      setInviteEmail('');
    } catch (err: unknown) {
      const message = (err as { response?: { data?: { error?: string } } }).response?.data?.error;
      setInviteError(message || 'Не удалось отправить приглашение.');
    } finally {
      setInviting(false);
    }
  };

  return (
    <div className="min-h-screen bg-blue-50 flex flex-col">
      <header className="bg-white shadow-sm border-b px-6 py-4">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4 min-w-0">
            <Link to="/boards" className="text-gray-600 hover:text-blue-600 font-medium whitespace-nowrap">← Все доски</Link>
            <h1 className="text-2xl font-bold text-gray-800 truncate">{board.title}</h1>
          </div>
          <button onClick={() => void openMembers()} className="px-4 py-2 bg-white border border-gray-300 rounded-md hover:bg-gray-50 text-gray-700 whitespace-nowrap">
            Участники
          </button>
        </div>
      </header>

      <DndContext sensors={sensors} collisionDetection={columnCollisionDetection}
        onDragStart={handleDragStart} onDragOver={handleDragOver} onDragEnd={handleDragEnd}
        onDragCancel={() => { setActiveCard(null); setActiveColumn(null); onReload(); }}>
        <main className="flex-1 overflow-x-auto p-6">
          <SortableContext items={sortedColumns.map((column) => column.id)} strategy={horizontalListSortingStrategy}>
            <div className="flex space-x-4 items-start">
              {sortedColumns.map((column) => (
                <SortableColumn key={column.id} column={column}
                  onAddCard={(columnId) => openCreationModal({ kind: 'card', columnId })}
                  onDeleteCard={setDeleteCardId} />
              ))}
              <button onClick={() => openCreationModal({ kind: 'column' })}
                className="flex-shrink-0 w-80 h-12 bg-blue-100/50 hover:bg-blue-100 border-2 border-dashed border-blue-300 rounded-lg flex items-center justify-center text-blue-700 font-medium transition-colors">
                + Добавить колонку
              </button>
            </div>
          </SortableContext>
        </main>
        <DragOverlay>
          {activeCard && <div className="bg-white p-3 rounded shadow-lg rotate-2 opacity-90 w-64"><p className="text-sm text-gray-800">{activeCard.title}</p></div>}
          {activeColumn && <div className="bg-gray-100 p-4 rounded-lg shadow-xl opacity-90 w-80"><p className="font-semibold">{activeColumn.title}</p><p className="text-xs text-gray-500 mt-2">{activeColumn.cards.length} карточек</p></div>}
        </DragOverlay>
      </DndContext>

      {creationModal && (
        <Modal title={creationModal.kind === 'column' ? 'Новая колонка' : 'Новая карточка'} onClose={() => setCreationModal(null)}>
          <form onSubmit={submitCreation} className="space-y-4">
            <label className="block text-sm font-medium text-gray-700">
              Название
              <input autoFocus maxLength={120} value={creationTitle} onChange={(event) => setCreationTitle(event.target.value)}
                placeholder={creationModal.kind === 'column' ? 'Например, В работе' : 'Например, Подготовить макет'}
                className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500" />
            </label>
            {creationError && <p className="text-sm text-red-600">{creationError}</p>}
            <div className="flex justify-end gap-2">
              <button type="button" onClick={() => setCreationModal(null)} className="px-4 py-2 rounded-md bg-gray-100 hover:bg-gray-200">Отмена</button>
              <button disabled={saving} type="submit" className="px-4 py-2 rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50">{saving ? 'Создание…' : 'Создать'}</button>
            </div>
          </form>
        </Modal>
      )}

      {deleteCardId && (
        <Modal title="Удалить карточку?" onClose={() => setDeleteCardId(null)}>
          <p className="text-sm text-gray-600">Карточка будет удалена без возможности восстановления.</p>
          <div className="mt-5 flex justify-end gap-2">
            <button onClick={() => setDeleteCardId(null)} className="px-4 py-2 rounded-md bg-gray-100 hover:bg-gray-200">Отмена</button>
            <button onClick={() => void confirmDeleteCard()} className="px-4 py-2 rounded-md bg-red-600 text-white hover:bg-red-700">Удалить</button>
          </div>
        </Modal>
      )}

      {showMembers && (
        <Modal title="Участники доски" onClose={() => setShowMembers(false)}>
          <div className="space-y-3 max-h-56 overflow-y-auto">
            {members.map((member) => (
              <div key={member.user_id} className="flex items-center justify-between rounded-md bg-gray-50 px-3 py-2">
                <span className="text-sm text-gray-800">{member.email}</span>
                <span className="text-xs text-gray-500">{member.role === 'admin' || member.role === 'owner' ? 'Администратор' : 'Участник'}</span>
              </div>
            ))}
            {members.length === 0 && <p className="text-sm text-gray-500">Нет данных об участниках.</p>}
          </div>
          {canInvite && (
            <form onSubmit={submitInvitation} className="mt-5 border-t pt-4 space-y-3">
              <h3 className="font-semibold text-gray-800">Пригласить по электронной почте</h3>
              <input type="email" required value={inviteEmail} onChange={(event) => setInviteEmail(event.target.value)} placeholder="user@example.com"
                className="w-full rounded-md border border-gray-300 px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500" />
              {inviteError && <p className="text-sm text-red-600">{inviteError}</p>}
              {inviteSuccess && <p className="text-sm text-green-700">{inviteSuccess}</p>}
              <div className="flex justify-end"><button disabled={inviting} type="submit" className="px-4 py-2 rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50">{inviting ? 'Отправка…' : 'Создать приглашение'}</button></div>
              <p className="text-xs text-gray-500">Приглашение появится в аккаунте с этим адресом. Автоматическая отправка email пока не настроена.</p>
            </form>
          )}
        </Modal>
      )}
    </div>
  );
};

const Modal = ({ title, onClose, children }: { title: string; onClose: () => void; children: ReactNode }) => (
  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
    <section role="dialog" aria-modal="true" aria-label={title} className="w-full max-w-md rounded-xl bg-white p-6 shadow-2xl">
      <div className="mb-4 flex items-center justify-between gap-4"><h2 className="text-xl font-bold text-gray-900">{title}</h2><button onClick={onClose} aria-label="Закрыть" className="text-2xl leading-none text-gray-400 hover:text-gray-700">×</button></div>
      {children}
    </section>
  </div>
);
