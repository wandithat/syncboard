import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import type {
  DragStartEvent,
  DragOverEvent,
  DragEndEvent,
} from '@dnd-kit/core';
import { SortableContext, horizontalListSortingStrategy } from '@dnd-kit/sortable';
import { boardsApi } from '../api/boards';
import { columnsApi } from '../api/columns';
import { cardsApi } from '../api/cards';
import type { Board, Card, ColumnWithCards } from '../types';
import { SortableColumn } from '../components/SortableColumn';
import { useBoardState } from '../hooks/useBoardState';

interface BoardData extends Board {
  columns: ColumnWithCards[];
}

export const BoardView = () => {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(true);
  const [initialBoard, setInitialBoard] = useState<BoardData | null>(null);

  const loadBoard = useCallback(async () => {
    if (!id) return;

    try {
      const boardData = await boardsApi.get(id);
      const columns = await columnsApi.getByBoard(id);

      const columnsWithCards: ColumnWithCards[] = await Promise.all(
        columns.map(async (col) => {
          const cards = await cardsApi.getByColumn(col.id);
          return { ...col, cards };
        })
      );

      setInitialBoard({ ...boardData, columns: columnsWithCards });
    } catch (err) {
      console.error('Ошибка загрузки доски:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadBoard();
  }, [loadBoard]);

  if (loading) {
    return <div className="flex justify-center items-center h-screen">Загрузка...</div>;
  }

  if (!initialBoard) {
    return <div className="flex justify-center items-center h-screen">Доска не найдена</div>;
  }

  return <BoardContent initialBoard={initialBoard} onReload={loadBoard} />;
};

interface BoardContentProps {
  initialBoard: BoardData;
  onReload: () => void;
}

const BoardContent = ({ initialBoard, onReload }: BoardContentProps) => {
  const {
    board,
    moveCardWithinColumn,
    moveCardBetweenColumns,
    addCardToColumn,
    removeCardFromColumn,
  } = useBoardState(initialBoard);

  const [activeCard, setActiveCard] = useState<Card | null>(null);

  // Сенсоры: мышь с задержкой, чтобы клик не начинал drag
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5, // 5px движения для старта drag
      },
    })
  );

  // Найти колонку по ID карточки
  const findColumnByCardId = (cardId: string): ColumnWithCards | undefined => {
    return board.columns.find((col) => col.cards.some((c) => c.id === cardId));
  };

  // === DRAG START ===
  const handleDragStart = (event: DragStartEvent) => {
    const { active } = event;
    if (active.data.current?.type === 'card') {
      setActiveCard(active.data.current.card);
    }
  };

  // === DRAG OVER: live-обновление при перемещении ===
  const handleDragOver = (event: DragOverEvent) => {
    const { active, over } = event;
    if (!over) return;

    const activeId = String(active.id);
    const overId = String(over.id);

    if (activeId === overId) return;

    // Работаем только с карточками
    if (active.data.current?.type !== 'card') return;

    const activeColumn = findColumnByCardId(activeId);
    
    // over может быть карточкой или самой колонкой (если бросаем в пустую)
    const overColumn =
      findColumnByCardId(overId) || board.columns.find((c) => c.id === overId);

    if (!activeColumn || !overColumn) return;

    // Если карточка перемещается в другую колонку
    if (activeColumn.id !== overColumn.id) {
      const overIndex = overColumn.cards.findIndex((c) => c.id === overId);
      
      moveCardBetweenColumns(
        activeId,
        activeColumn.id,
        overColumn.id,
        overIndex >= 0 ? overIndex : overColumn.cards.length
      );
    }
  };

  // === DRAG END: финализация и сохранение ===
  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveCard(null);

    if (!over) return;

    const activeId = String(active.id);
    const overId = String(over.id);

    if (activeId === overId) return;

    if (active.data.current?.type !== 'card') return;

    const column = findColumnByCardId(activeId);
    if (!column) return;

    const oldIndex = column.cards.findIndex((c) => c.id === activeId);
    const newIndex = column.cards.findIndex((c) => c.id === overId);

    // Перемещение внутри колонки
    if (oldIndex !== newIndex && newIndex !== -1) {
      moveCardWithinColumn(column.id, oldIndex, newIndex);
    }

    // Сохраняем на сервер
    const card = column.cards.find((c) => c.id === activeId);
    if (!card) return;

    try {
      await cardsApi.move(activeId, card.column_id, card.position);
    } catch (err) {
      console.error('Ошибка сохранения:', err);
      // Откат: перезагружаем доску с сервера
      onReload();
      alert('Не удалось сохранить изменения');
    }
  };

  // === Добавление колонки ===
  const handleAddColumn = async () => {
    const title = prompt('Название новой колонки:');
    if (title?.trim()) {
      try {
        await columnsApi.create(board.id, title.trim());
        onReload();
      } catch (err) {
        alert('Ошибка создания колонки');
      }
    }
  };

  // === Добавление карточки ===
  const handleAddCard = async (columnId: string) => {
    const title = prompt('Название карточки:');
    if (title?.trim()) {
      try {
        const newCard = await cardsApi.create(columnId, title.trim());
        addCardToColumn(columnId, newCard);
      } catch (err) {
        alert('Ошибка создания карточки');
      }
    }
  };

  // === Удаление карточки ===
  const handleDeleteCard = async (cardId: string) => {
    if (confirm('Удалить эту карточку?')) {
      try {
        await cardsApi.delete(cardId);
        removeCardFromColumn(cardId);
      } catch (err) {
        alert('Ошибка удаления');
      }
    }
  };

  return (
    <div className="min-h-screen bg-blue-50 flex flex-col">
      <header className="bg-white shadow-sm border-b px-6 py-4">
        <div className="flex items-center space-x-4">
          <Link to="/boards" className="text-gray-600 hover:text-blue-600 font-medium">
            ← Все доски
          </Link>
          <h1 className="text-2xl font-bold text-gray-800">{board.title}</h1>
        </div>
      </header>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
      >
        <main className="flex-1 overflow-x-auto p-6">
          <SortableContext
            items={board.columns.map((c) => c.id)}
            strategy={horizontalListSortingStrategy}
          >
            <div className="flex space-x-4 items-start">
              {board.columns
                .sort((a, b) => a.position - b.position)
                .map((column) => (
                  <SortableColumn
                    key={column.id}
                    column={column}
                    onAddCard={handleAddCard}
                    onDeleteCard={handleDeleteCard}
                  />
                ))}

              <button
                onClick={handleAddColumn}
                className="flex-shrink-0 w-80 h-12 bg-blue-100/50 hover:bg-blue-100 border-2 border-dashed border-blue-300 rounded-lg flex items-center justify-center text-blue-700 font-medium transition-colors"
              >
                + Добавить колонку
              </button>
            </div>
          </SortableContext>
        </main>

        {/* DragOverlay: "призрак" карточки, следующий за курсором */}
        <DragOverlay>
          {activeCard && (
            <div className="bg-white p-3 rounded shadow-lg rotate-3 opacity-90">
              <p className="text-sm text-gray-800">{activeCard.title}</p>
            </div>
          )}
        </DragOverlay>
      </DndContext>
    </div>
  );
};
