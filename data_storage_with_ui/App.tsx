import { useState, useEffect } from 'react';

function App() {
    const [activeTab, setActiveTab] = useState('list');

    const [keys, setKeys] = useState([]);
    const [selectedKey, setSelectedKey] = useState('');
    const [content, setContent] = useState('');
    const [error, setError] = useState('');

    const [keyInput, setKeyInput] = useState('');
    const [jsonInput, setJsonInput] = useState('');

    // load list
    const loadKeys = async () => {
        try {
            const response = await fetch('http://localhost:8080/list');
            const data = await response.json();
            if (data && data.keys) {
                setKeys(data.keys);
            } else {
                setError('Could not load keys.');
            }
        } catch (err: any) {
            setError('Error: ' + err.message);
        }
    };

    // refresh list on active tab
    useEffect(() => {
        if (activeTab === 'list') {
            loadKeys();
        }
    }, [activeTab]);

    // load specific key
    const loadKeyContent = async (key: string) => {
        try {
            const response = await fetch(`http://localhost:8080/data?key=${encodeURIComponent(key)}`);
            const data = await response.json();
            if (data && data.json) {
                // format
                setContent(JSON.stringify(data.json, null, 2));
            } else {
                setContent('No content found for this key.');
            }
        } catch (err: any) {
            setContent('Error: ' + err.message);
        }
    };

    // handle sending data
    const sendData = async () => {
        // validate
        let parsedJson;
        try {
            parsedJson = JSON.parse(jsonInput);
        } catch (err) {
            alert('Invalid JSON data');
            return;
        }

        try {
            const response = await fetch(`http://localhost:8080/data?key=${encodeURIComponent(keyInput)}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ json: parsedJson }),
            });
            const data = await response.json();
            if (data.message) {
                alert(data.message);
                // clear inputs
                setKeyInput('');
                setJsonInput('');
            } else if (data.error) {
                alert(data.error);
            }
        } catch (err: any) {
            alert('Error: ' + err.message);
        }
    };

    return (
        <div style={{minWidth: '1000px', width: '80%', margin: '0 10%', padding: '20px', fontFamily: 'sans-serif'}}>
            <div style={{marginBottom: '20px'}}>
                <button onClick={() => setActiveTab('list')} style={{marginRight: '10px'}}>
                    List Data
                </button>
                <button onClick={() => setActiveTab('send')}>Send Data</button>
            </div>
            <hr/>

            {activeTab === 'list' && (
                <div>
                    <h2>Existing Keys</h2>
                    {error && <p style={{color: 'red'}}>{error}</p>}
                    <ul>
                        {keys.map((k) => (
                            <li key={k} style={{marginBottom: '5px'}}>
                                {k}{' '}
                                <button
                                    onClick={() => {
                                        setSelectedKey(k);
                                        loadKeyContent(k);
                                    }}
                                >
                                    Load Content
                                </button>
                            </li>
                        ))}
                    </ul>
                    {selectedKey && (
                        <div style={{marginTop: '20px'}}>
                            <h3>Content for key: {selectedKey}</h3>
                            <pre style={{background: '#333', color: '#fff', padding: '10px'}}>{content}</pre>
                        </div>
                    )}
                </div>
            )}

            {activeTab === 'send' && (
                <div>
                    <h2>Send Data</h2>
                    <div style={{marginBottom: '10px'}}>
                        <label>
                            Key:
                            <input
                                type="text"
                                value={keyInput}
                                onChange={(e) => setKeyInput(e.target.value)}
                                style={{marginLeft: '10px'}}
                            />
                        </label>
                    </div>
                    <div style={{marginBottom: '10px'}}>
                        <label>
                            JSON Data:
                            <textarea
                                value={jsonInput}
                                onChange={(e) => setJsonInput(e.target.value)}
                                style={{width: '100%', height: '150px', marginTop: '5px'}}
                            />
                        </label>
                    </div>
                    <button onClick={sendData}>Send</button>
                </div>
            )}
        </div>
    );
}

export default App;
