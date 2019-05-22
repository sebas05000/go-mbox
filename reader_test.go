package mbox

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"testing"
)

const mboxWithOneMessage = `From herp.derp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxWithOneMessageMissingHeaders = `From herp.derp@example.com Thu Jan  1 00:00:01 2015
This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxWithThreeMessages = `From herp.derp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.

From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: derp.herp@example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Another line.

Bye.

From bernd.lauert@example.com Thu Jan  3 00:00:01 2015
From: bernd.lauert@example.com (Bernd Lauert)
Date: Thu, 03 Jan 2015 00:00:01 +0100
Subject: A last test

This is the last simple test.

Bye.
`

const mboxWithStartingLF = `
From herp.derp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.

From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: derp.herp@example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Another line.

Bye.

From bernd.lauert@example.com Thu Jan  3 00:00:01 2015
From: bernd.lauert@example.com (Bernd Lauert)
Date: Thu, 03 Jan 2015 00:00:01 +0100
Subject: A last test

This is the last simple test.

Bye.
`

const mboxWithThreeMessagesMalformedButValid = `From herp.derp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: derp.herp@example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Another line.

Bye.

From bernd.lauert@example.com Thu Jan  3 00:00:01 2015
From: bernd.lauert@example.com (Bernd Lauert)
Date: Thu, 03 Jan 2015 00:00:01 +0100
Subject: A last test

This is the last simple test.

Bye.
`

const mboxWithOneMessageMissingSeparator = `From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxFirstMessage = `From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

type tsmInput struct {
	data  string
	atEOF bool
}

type tsmExpected struct {
	advance     int
	token       string
	yieldsError bool
}

func testScanMessage(t *testing.T, input *tsmInput, expected *tsmExpected) {
	advance, token, err := scanMessage([]byte(input.data), input.atEOF)

	if err == nil && expected.yieldsError {
		t.Errorf("unexpected success")
	}
	if err != nil && !expected.yieldsError {
		t.Errorf("unexpected error: %v", err)
	}
	if advance != expected.advance {
		t.Errorf("unexpected advance: %d", advance)
	}
	if string(token) != expected.token {
		t.Errorf("unexpected token: %q", token)
	}
}

func TestScanMessageMboxEmptyAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  "",
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithOneMessageAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessage,
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     274,
		token:       mboxFirstMessage,
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithOneMessageMissingSeparatorAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessageMissingSeparator,
	}

	expected := &tsmExpected{
		yieldsError: true,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithThreeMessages(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  mboxWithThreeMessages,
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     275,
		token:       mboxFirstMessage,
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageWithThreeMessagesMalformedButValid(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  mboxWithThreeMessagesMalformedButValid,
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     274,
		token:       mboxFirstMessage,
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageIncompleteRecord(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  mboxWithOneMessage[:100],
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageVeryShortIncompleteRecord(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  "From",
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageOnlySeperatorAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessage[:55],
	}

	expected := &tsmExpected{
		yieldsError: true,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithOneMessageWithoutNewlineAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessage[:len(mboxWithOneMessage)-1],
	}

	expected := &tsmExpected{
		yieldsError: true,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func testMboxMessage(t *testing.T, mbox string, count int) {
	b := bytes.NewBufferString(mbox)
	m := NewReader(b)

	for i := 0; i < count; i++ {
		r, err := m.NextMessage()
		if err != nil {
			t.Fatalf("Unexpected error after NextMessage(): %v", err)
		}

		var text bytes.Buffer
		_, err = text.ReadFrom(r)
		if err != nil {
			t.Errorf("Unexpected error reading message body: %v", err)
		}
		if i == 0 && text.String() != mboxFirstMessage {
			t.Errorf("Expected:\n %q\ngot\n%q", mboxFirstMessage, text.String())
		}
	}

	if _, err := m.NextMessage(); err != io.EOF {
		t.Fatalf("Unexpected error after NextMessage(): %v", err)
	}
}

func TestMboxMessageWithOneMessage(t *testing.T) {
	testMboxMessage(t, mboxWithOneMessage, 1)
}

func TestMboxMessageWithThreeMessages(t *testing.T) {
	testMboxMessage(t, mboxWithThreeMessages, 3)
}

func TestMboxMessageWithStartingLF(t *testing.T) {
	testMboxMessage(t, mboxWithStartingLF, 3)
}

func TestMboxMessageWithThreeMessagesMalformedButValid(t *testing.T) {
	testMboxMessage(t, mboxWithThreeMessagesMalformedButValid, 3)
}

func testMboxMessageInvalid(t *testing.T, mbox string) {
	b := bytes.NewBufferString(mbox)
	m := NewReader(b)

	if _, err := m.NextMessage(); err == nil {
		t.Errorf("Missing error after Next(): %v", err)
	}
}

func TestMboxMessageWithOneMessageMissingSeparator(t *testing.T) {
	testMboxMessageInvalid(t, mboxWithOneMessageMissingSeparator)
}

func TestScanMessageWithBoundaries(t *testing.T) {
	sourceData := `
From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test
Content-Type: multipart/alternative;
        boundary=Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0

This is a test of boundaries.  Don't accept a new email via \nFrom until the boundary is done!'

And, by the way, this is how a "From" line is escaped in mboxo format:
From Herp Derp with love.

From Herp Derp with love.

Bye.
--Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0--

From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is the second email in a test of boundaries.
`
	expected := []string{
		"This is a test of boundaries.  Don't accept a new email via \\nFrom until the boundary is done!'\n\nAnd, by the way, this is how a \"From\" line is escaped in mboxo format:\nFrom Herp Derp with love.\n\nFrom Herp Derp with love.\n\nBye.\n--Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0--\n",
		"This is the second email in a test of boundaries.\n",
	}
	b := bytes.NewBufferString(sourceData)
	m := NewReader(b)

	for i := range expected {
		r, err := m.NextMessage()
		if err != nil {
			t.Fatalf("Unexpected error after NextMessage(): %v", err)
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			t.Fatalf("mail.ReadMessage() = %v", err)
		}

		var body bytes.Buffer
		_, err = body.ReadFrom(msg.Body)
		if err != nil {
			t.Errorf("%d - Unexpected error reading message body: %v", i, err)
			continue
		}
		if body.String() != expected[i] {
			t.Errorf("%d - Expected:\n %q\ngot\n%q", i, expected[i], body.String())
		}
	}

	if _, err := m.NextMessage(); err != io.EOF {
		t.Errorf("Next() succeeded")
	}
}

func TestScanMessageWithTextBoundary(t *testing.T) {
	sourceData := `
From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test
Content-Type: text/html; charset="utf-8";
 boundary="monkey_d3df4dc8-da5e-47dd-be15-f19c5ed55194"

This is a test of boundaries.  Don't accept a new email via \nFrom until the boundary is done!'

And, by the way, this is how a "From" line is escaped in mboxo format:

Bye.

From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is the second email in a test of boundaries.
`
	expected := []string{
		"This is a test of boundaries.  Don't accept a new email via \\nFrom until the boundary is done!'\n\nAnd, by the way, this is how a \"From\" line is escaped in mboxo format:\n\nBye.\n",
		"This is the second email in a test of boundaries.\n",
	}
	b := bytes.NewBufferString(sourceData)
	m := NewReader(b)

	for i := range expected {
		r, err := m.NextMessage()
		if err != nil {
			t.Fatalf("Unexpected error after NextMessage(): %v", err)
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			t.Fatalf("mail.ReadMessage() = %v", err)
		}

		var body bytes.Buffer
		_, err = body.ReadFrom(msg.Body)
		if err != nil {
			t.Errorf("%d - Unexpected error reading message body: %v", i, err)
			continue
		}
		if body.String() != expected[i] {
			t.Errorf("%d - Expected:\n %q\ngot\n%q", i, expected[i], body.String())
		}
	}

	if _, err := m.NextMessage(); err != io.EOF {
		t.Errorf("Next() succeeded")
	}
}

func TestScanMessageWithBoundarySemicolon(t *testing.T) {
	mbox := `From notifications@github.com Tue Jun  7 05:46:46 2016
From: Sender <notifications@github.com>
To: foo/bar <bar@noreply.github.com>
Message-ID: <foo/bar/1/0@github.com>
Subject: Re: [foo/bar] [question] Baz? (#1)
Content-Type: multipart/alternative; boundary="--==_mimepart_5755da228145a_38da3facdf97329c42987b";
MIME-Version: 1.0


----==_mimepart_5755da228145a_38da3facdf97329c42987b
Content-Type: text/plain; charset="UTF-8"
Content-Transfer-Encoding: 8bit

Blah blah

----==_mimepart_5755da228145a_38da3facdf97329c42987b
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: 8bit

<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<p>Blah blah</p>

----==_mimepart_5755da228145a_38da3facdf97329c42987b--

From notifications@github.com Tue Jun  7 05:52:15 2016
From: Author <notifications@github.com>
To: frob/blab <blab@noreply.github.com>
Message-ID: <frob/blab/1/0@github.com>
Subject: Re: [frob/blab] [question] Bling? (#1)
Content-Type: multipart/alternative; boundary="--==_mimepart_5755db739a819_79783f996b0172c04025ee";
MIME-Version: 1.0


----==_mimepart_5755db739a819_79783f996b0172c04025ee
Content-Type: text/plain; charset="UTF-8"
Content-Transfer-Encoding: 8bit

Blah blah

----==_mimepart_5755db739a819_79783f996b0172c04025ee
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: 8bit

<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<p>Blah blah</p>

----==_mimepart_5755db739a819_79783f996b0172c04025ee--

`
	expected := 2

	b := bytes.NewBufferString(mbox)
	m := NewReader(b)

	parsedMessages := 0
	for {
		_, err := m.NextMessage()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("m.NextMessaage() = %v", err)
		}
		parsedMessages += 1
	}

	if parsedMessages != expected {
		t.Errorf("Expected: %d; got: %d", expected, parsedMessages)
	}
}

func ExampleReader() {
	r := strings.NewReader(`From herp.derp@example.com Thu Jan  1 00:00:01 2015
From: herp.derp@example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

CU.

From derp.herp@example.com Thu Jan  1 00:00:01 2015
From: derp.herp@example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Bye.
`)

	mr := NewReader(r)
	for {
		r, err := mr.NextMessage()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Print("Oops, something went wrong!", err)
			return
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			fmt.Print("Oops, something went wrong!", err)
			return
		}

		fmt.Printf("Message from %v\n", msg.Header.Get("From"))
	}

	// Output:
	// Message from herp.derp@example.com (Herp Derp)
	// Message from derp.herp@example.com (Derp Herp)
}