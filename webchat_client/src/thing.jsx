import {useState} from 'react'

export default function MessageBox({socket}) {
	const [postContent, setPost] = useState("");

	function handleSubmit(e) {
		e.preventDefault();
		// const form = e.target;
		// const formData = new FormData(form);

		// const formJson = Object.fromEntries(formData.entries());
		if(postContent == "") {
			return
		}
		// socket.current.send("<" + user + "> " + postContent)
		socket.current.send(JSON.stringify({msg: postContent, kind: 1}))
		setPost("")
	}

	return (
		<form name="chat" method="post" onSubmit={handleSubmit}>
			<input
				type="text"
				value={postContent}
				onChange={e => setPost(e.target.value)}
				placeholder="type a message!"
				size={40}
				maxlength={255}
			/>
			<br />
		</form>
	)
}
