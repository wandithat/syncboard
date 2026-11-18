import { useEffect, useState } from 'react';
import { useEffect, useState } from 'react';
import type { BoardWithDetails, Card } from '../types';

export const useBoardState = (initialBoard: BoardWithDetails) => {
  const [board, setBoard] = useState<BoardWithDetails>(initialBoard);
  useEffect(() => {
    setBoard(initialBoard);
  }, [initialBoard]);

  const reorderColumns = (orderedIds: string[]) => {
    setBoard((prev) => {
      const byId = new Map(prev.columns.map((column) => [column.id, column]));
      const columns = orderedIds
        .map((id, position) => {
          const column = byId.get(id);
          return column ? { ...column, position } : undefined;
        })
        .filter((column): column is NonNullable<typeof column> => Boolean(column));
      return { ...prev, columns };
    });
  };

  // Переместить карточку в пределах одной колонки
  const moveCardWithinColumn = (
    columnId: string,
    fromIndex: number,
    toIndex: number
  ) => {
    setBoard((prev) => ({
      ...prev,
      columns: prev.columns.map((col) => {
        if (col.id !== columnId) return col;
        
        const cards = [...col.cards];
        if (fromIndex < 0 || fromIndex >= cards.length || toIndex < 0 || toIndex >= cards.length) return col;
        if (fromIndex < 0 || fromIndex >= cards.length || toIndex < 0 || toIndex >= cards.length) return col;
        const [moved] = cards.splice(fromIndex, 1);
        if (!moved) return col;
        if (!moved) return col;
        cards.splice(toIndex, 0, moved);
        
        // Обновляем позиции всех карточек в колонке
        const updatedCards = cards.map((card, idx) => ({
          ...card,
          position: idx,
        }));
        
        return { ...col, cards: updatedCards };
      }),
    }));
  };

  // Переместить карточку между колонками
  const moveCardBetweenColumns = (
    cardId: string,
    fromColumnId: string,
    toColumnId: string,
    toIndex: number
  ) => {
    setBoard((prev) => {
      const fromColumn = prev.columns.find((c) => c.id === fromColumnId);
      if (!fromColumn) return prev;

      const card = fromColumn.cards.find((c) => c.id === cardId);
      if (!card) return prev;

      return {
        ...prev,
        columns: prev.columns.map((col) => {
          // Удаляем из старой колонки
          if (col.id === fromColumnId) {
            const cards = col.cards.filter((c) => c.id !== cardId);
            const updatedCards = cards.map((c, idx) => ({ ...c, position: idx }));
            return { ...col, cards: updatedCards };
          }
          
          // Добавляем в новую колонку
          if (col.id === toColumnId) {
            const cards = [...col.cards];
            const newCard = { ...card, column_id: toColumnId };
            cards.splice(toIndex, 0, newCard);
            const updatedCards = cards.map((c, idx) => ({ ...c, position: idx }));
            return { ...col, cards: updatedCards };
          }
          
          return col;
        }),
      };
    });
  };

  // Добавить карточку в колонку
  const addCardToColumn = (columnId: string, card: Card) => {
    setBoard((prev) => ({
      ...prev,
      columns: prev.columns.map((col) => {
        if (col.id !== columnId) return col;
        return { ...col, cards: [...col.cards, { ...card, position: col.cards.length }] };
      }),
    }));
  };

  // Удалить карточку из колонки
  const removeCardFromColumn = (cardId: string) => {
    setBoard((prev) => ({
      ...prev,
      columns: prev.columns.map((col) => {
        const cards = col.cards.filter((c) => c.id !== cardId);
        if (cards.length === col.cards.length) return col;
        
        const updatedCards = cards.map((c, idx) => ({ ...c, position: idx }));
        return { ...col, cards: updatedCards };
      }),
    }));
  };

  return {
    board,
    setBoard,
    reorderColumns,
    moveCardWithinColumn,
    moveCardBetweenColumns,
    addCardToColumn,
    removeCardFromColumn,
  };
};