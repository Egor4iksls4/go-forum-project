import { useState, useEffect, useRef } from "react";
import "./News.css";

export default function News() {
    const [news, setNews] = useState(() => {
        const saved = localStorage.getItem('news');
        return saved ? JSON.parse(saved) : [];
    });
    const [newTitle, setNewTitle] = useState('');
    const [newContent, setNewContent] = useState('');
    const [editingId, setEditingId] = useState('');
    const [editTitle, setEditTitle] = useState('');
    const [editContent, setEditContent] = useState('');
    const [showAddForm, setShowAddForm] = useState(false);
    const isAuth = localStorage.getItem('isAuth') === 'true';
    const formRef = useRef(null);

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (formRef.current && !formRef.current.contains(event.target)) {
                setShowAddForm(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    useEffect(() => {
        localStorage.setItem('news', JSON.stringify(news));
    }, [news]);

    const addNews = (e) => {
        e.preventDefault();
        if (!newTitle || !newContent) return;
        const updatedNews = [...news, { id: Date.now(), title: newTitle, content: newContent }];
        setNews(updatedNews);
        setNewTitle('');
        setNewContent('');
        setShowAddForm(false);
    };

    const deleteNews = (id) => {
        const confirmDelete = window.confirm('Удалить новость?');
        if (confirmDelete) {
            setNews(news.filter(item => item.id !== id));
        }
    };

    const startEdit = (newsItem) => {
        setEditingId(newsItem.id);
        setEditTitle(newsItem.title);
        setEditContent(newsItem.content);
    };

    const saveEdit = (e) => {
        e.preventDefault();
        const updatedNews = news.map(item => 
            item.id === editingId ? { ...item, title: editTitle, content: editContent } : item
        );
        setNews(updatedNews);
        setEditingId(null);
    };

    return (
        <main className="news-main">
            <h1 className="news-header">Темы</h1>
            {isAuth && (
                <button className="add-button" onClick={() => setShowAddForm(!showAddForm)}>
                    Добавить тему
                </button>
            )}
            {showAddForm && (
                <div className="add-news-overlay">
                    <div ref={formRef} className="add-news">
                        <form onSubmit={addNews} className="form-style">
                            <input
                                type="text"
                                value={newTitle}
                                onChange={(e) => setNewTitle(e.target.value)}
                                placeholder="Заголовок"
                                className="input-style"
                                required
                            />
                            <textarea
                                value={newContent}
                                onChange={(e) => setNewContent(e.target.value)}
                                placeholder="Содержание"
                                className="textarea-style"
                                required
                            />
                            <div className="form-buttons">
                                <button type="submit">Добавить новость</button>
                                <button
                                    type="button"
                                    onClick={() => setShowAddForm(false)}
                                    className="cancel-button"
                                >
                                    Отмена
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
            <div className="news-list">
                {news.map(item => (
                    <div key={item.id} className="news-items">
                        {editingId === item.id ? (
                            <form onSubmit={saveEdit}>
                                <input 
                                    type="text"
                                    value={editTitle}
                                    onChange={(e) => setEditTitle(e.target.value)}
                                    className="input-style"
                                />
                                <textarea 
                                    value={editContent}
                                    onChange={(e) => setEditContent(e.target.value)}
                                    className="textarea-style"
                                />
                                <div className="redact-buttons">
                                    <button type="submit">Сохранить</button>
                                    <button type="button" onClick={() => setEditingId(null)}>Отмена</button>
                                </div>
                            </form>
                        ) : (
                            <>
                                <h3>{item.title}</h3>
                                <p>{item.content}</p>
                                {isAuth && (
                                    <div className="redact-buttons">
                                        <button onClick={() => startEdit(item)}>
                                            Редактировать
                                        </button>
                                        <button onClick={() => deleteNews(item.id)}>Удалить</button>
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                ))}
            </div>
        </main>
    );
}