import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { useState } from 'react';

const registerSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(8, 'Минимум 8 символов'),
});

type RegisterForm = z.infer<typeof registerSchema>;

export const Register = () => {
  const navigate = useNavigate();
  const registerUser = useAuthStore((state) => state.register);
  const [error, setError] = useState('');

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterForm) => {
    try {
      setError('');
      await registerUser(data.email, data.password);
      navigate('/login'); // После регистрации отправляем на вход
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка регистрации');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-xl shadow-lg">
        <h2 className="text-3xl font-bold text-center text-gray-900">Регистрация</h2>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          {error && <div className="bg-red-50 text-red-600 p-3 rounded-md text-sm">{error}</div>}
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Email</label>
              <input {...register('email')} type="email" className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md" />
              {errors.email && <p className="mt-1 text-sm text-red-500">{errors.email.message}</p>}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Пароль</label>
              <input {...register('password')} type="password" className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md" />
              {errors.password && <p className="mt-1 text-sm text-red-500">{errors.password.message}</p>}
            </div>
          </div>
          <button type="submit" disabled={isSubmitting} className="w-full py-2.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
            {isSubmitting ? 'Создание...' : 'Зарегистрироваться'}
          </button>
          <div className="text-center">
            <Link to="/login" className="text-sm text-blue-600 hover:text-blue-500">Уже есть аккаунт? Войти</Link>
          </div>
        </form>
      </div>
    </div>
  );
};
