import {
  useSortable,
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { ColumnWithCards } from '../types';
import { SortableCard } from './SortableCard';

interface SortableColumnProps {
  column: ColumnWithCards;
  onAddCard: (columnId: string) => void;
  onDeleteCard: (cardId: string) => void;
}

export const SortableColumn = ({
  column,
  onAddCard,
  onDeleteCard,
}: SortableColumnProps) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: column.id,
    data: {
      type: 'column',
      column,
    },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  // ID карточек для SortableContext
  const cardIds = column.cards.map((card) => card.id);

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="flex-shrink-0 w-80 bg-gray-100 rounded-lg p-3 shadow-sm"
    >
      {/* Заголовок колонки — перетаскиваемый */}
      <div
        {...attributes}
        {...listeners}
        className="font-semibold text-gray-700 mb-3 px-1 cursor-grab active:cursor-grabbing select-none"
      >
        {column.title}
      </div>

      {/* Список карточек */}
      <SortableContext items={cardIds} strategy={verticalListSortingStrategy}>
        <div className="space-y-2 min-h-[50px]">
          {column.cards
            .sort((a, b) => a.position - b.position)
            .map((card) => (
              <SortableCard key={card.id} card={card} onDelete={onDeleteCard} />
            ))}
        </div>
      </SortableContext>

      {/* Кнопка добавления карточки */}
      <button
        onClick={() => onAddCard(column.id)}
        className="w-full mt-3 py-2 text-sm text-gray-600 hover:bg-gray-200 rounded-md transition-colors text-left px-2"
      >
        + Добавить карточку
      </button>
    </div>
  );
};
