import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { boardsApi } from '../api/boards';
import { columnsApi } from '../api/columns';
import { cardsApi } from '../api/cards';
import type { Board, Card, Column } from '../types';

interface ColumnWithCards extends Column {
  cards: Card[];
}

interface BoardData extends Board {
  columns: ColumnWithCards[];
}

export const BoardView = () => {
  const { id } = useParams<{ id: string }>();
  const [board, setBoard] = useState<BoardData | null>(null);
  const [loading, setLoading] = useState(true);

  const loadBoard = useCallback(async () => {
    if (!id) return;

    try {
      // 1. Загружаем доску
      const boardData = await boardsApi.get(id);

      // 2. Загружаем её колонки
      const columns = await columnsApi.getByBoard(id);

      // 3. Загружаем карточки каждой колонки (параллельно)
      const columnsWithCards: ColumnWithCards[] = await Promise.all(
        columns.map(async (col) => {
          const cards = await cardsApi.getByColumn(col.id);
          return { ...col, cards };
        })
      );

      setBoard({ ...boardData, columns: columnsWithCards });
    } catch (err) {
      console.error('Ошибка загрузки доски:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadBoard();
  }, [loadBoard]);

  const handleAddColumn = async () => {
    const title = prompt('Название новой колонки:');
    if (title?.trim() && board) {
      try {
        await columnsApi.create(board.id, title.trim());
        await loadBoard();
      } catch (err) {
        alert('Ошибка создания колонки');
      }
    }
  };

  const handleAddCard = async (columnId: string) => {
    const title = prompt('Название карточки:');
    if (title?.trim()) {
      try {
        await cardsApi.create(columnId, title.trim());
        await loadBoard();
      } catch (err) {
        alert('Ошибка создания карточки');
      }
    }
  };

  const handleDeleteCard = async (cardId: string) => {
    if (confirm('Удалить эту карточку?')) {
      try {
        await cardsApi.delete(cardId);
        await loadBoard();
      } catch (err) {
        alert('Ошибка удаления');
      }
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen text-gray-500">
        Загрузка доски...
      </div>
    );
  }

  if (!board) {
    return (
      <div className="flex justify-center items-center h-screen">
        Доска не найдена
      </div>
    );
  }

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

      <main className="flex-1 overflow-x-auto p-6">
        <div className="flex space-x-4 items-start">
          {board.columns
            .sort((a, b) => a.position - b.position)
            .map((column) => (
              <ColumnView
                key={column.id}
                column={column}
                onAddCard={() => handleAddCard(column.id)}
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
      </main>
    </div>
  );
};

/* ---------- Колонка ---------- */

interface ColumnViewProps {
  column: ColumnWithCards;
  onAddCard: () => void;
  onDeleteCard: (cardId: string) => void;
}

const ColumnView = ({ column, onAddCard, onDeleteCard }: ColumnViewProps) => {
  return (
    <div className="flex-shrink-0 w-80 bg-gray-100 rounded-lg p-3 shadow-sm">
      <h3 className="font-semibold text-gray-700 mb-3 px-1">{column.title}</h3>

      <div className="space-y-2 min-h-[50px]">
        {column.cards
          .sort((a, b) => a.position - b.position)
          .map((card) => (
            <div
              key={card.id}
              className="bg-white p-3 rounded shadow-sm hover:shadow-md cursor-pointer group relative"
            >
              <p className="text-sm text-gray-800 pr-6">{card.title}</p>
              <button
                onClick={() => onDeleteCard(card.id)}
                title="Удалить карточку"
                className="absolute top-2 right-2 text-gray-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity text-lg leading-none"
              >
                ×
              </button>
            </div>
          ))}
      </div>

      <button
        onClick={onAddCard}
        className="w-full mt-3 py-2 text-sm text-gray-600 hover:bg-gray-200 rounded-md transition-colors text-left px-2"
      >
        + Добавить карточку
      </button>
    </div>
  );
};