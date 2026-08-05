import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Card } from '../types';

interface SortableCardProps {
  card: Card;
  onDelete: (cardId: string) => void;
}

export const SortableCard = ({ card, onDelete }: SortableCardProps) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: card.id,
    data: {
      type: 'card',
      card,
    },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="bg-white p-3 rounded shadow-sm hover:shadow-md cursor-grab active:cursor-grabbing group relative"
    >
      <div className="flex justify-between items-start">
        <p className="text-sm text-gray-800 pr-6 select-none">{card.title}</p>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(card.id);
          }}
          onPointerDown={(e) => e.stopPropagation()}
          title="Удалить карточку"
          className="absolute top-2 right-2 text-gray-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity text-lg leading-none"
        >
          ×
        </button>
      </div>
    </div>
  );
};
