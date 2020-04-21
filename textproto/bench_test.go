package textproto

import (
	"bufio"
	"net/textproto"
	"strings"
	"testing"
)

const testHeaderString1 = "Return-Path: <aaaaaaa@example.com>\r\n" +
	"Delivered-To: aaaaaaa@example.com\r\n" +
	"Received: from localhost (localhost [127.0.0.1])\r\n" +
	"    by example.com (Postfix) with ESMTP id 27A702E253\r\n" +
	"    for <aaaaaaa@example.com>; Fri, 11 Oct 2019 22:25:14 +0200 (CEST)\r\n" +
	"X-Virus-Scanned: Debian amavisd-new at example.com\r\n" +
	"X-Spam-Flag: NO\r\n" +
	"X-Spam-Score: -2.1\r\n" +
	"X-Spam-Level:\r\n" +
	"X-Spam-Status: No, score=-2.1 tagged_above=-9999 required=5\r\n" +
	"    tests=[BAYES_00=-1.9, DKIMWL_WL_HIGH=-0.001, DKIM_SIGNED=0.1,\r\n" +
	"    DKIM_VALID=-0.1, DKIM_VALID_AU=-0.1, DKIM_VALID_EF=-0.1,\r\n" +
	"    HTML_MESSAGE=0.001, SPF_HELO_NONE=0.001, SPF_PASS=-0.001]\r\n" +
	"    autolearn=ham autolearn_force=no\r\n" +
	"Received: from knopi.example.com ([127.0.0.1])\r\n" +
	"   by localhost (example.com [127.0.0.1]) (amavisd-new, port 10024)\r\n" +
	"   with ESMTP id 7VNTFhobvC5w for <aaaaaaa@example.com>;\r\n" +
	"   Fri, 11 Oct 2019 22:25:12 +0200 (CEST)\r\n" +
	"Received: from aaaaaaa.example.com (aaaaaaa.example.com [208.64.202.54])\r\n" +
	"   by example.com (Postfix) with ESMTPS id 498DF26A80\r\n" +
	"   for <aaaaaaa@example.com>; Fri, 11 Oct 2019 22:25:11 +0200 (CEST)\r\n" +
	"Authentication-Results: example.com;\r\n" +
	"    dkim=pass (1024-bit key; unprotected) header.d=example.com header.i=@example.com header.b=\"q+sL9woc\";\r\n" +
	"    dkim-atps=neutral\r\n" +
	"DKIM-Signature: v=1; a=rsa-sha256; q=dns/txt; c=relaxed/relaxed;\r\n" +
	"   d=example.com; s=smtp; h=Date:Message-Id:Content-Type:Subject:\r\n" +
	"   MIME-Version:Reply-To:From:To:Sender:Cc:Content-Transfer-Encoding:Content-ID:\r\n" +
	"   Content-Description:Resent-Date:Resent-From:Resent-Sender:Resent-To:Resent-Cc\r\n" +
	"   :Resent-Message-ID:In-Reply-To:References:List-Id:List-Help:List-Unsubscribe:\r\n" +
	"   List-Subscribe:List-Post:List-Owner:List-Archive;\r\n" +
	"   bh=oE7DVQcCo/SZapqnvRiosGrj0I7XI3GdKR8JEFu0l1U=; b=q+sL9wocDMrSmDdrErStDEN4AD\r\n" +
	"   tZQmJUGprHWQAuc4b+r5H/yKHqgGRMKm91qvxLfPtBbLIIAilDZk7Q6HAMko/qt9msj26eCO8a5/+\r\n" +
	"   QtT94z5b+p5OckmgpQTK9k5n3MlaHkLannUrMvr0+fKI/Nw5OMgmrlxzfFWS3gtri7ME=;\r\n" +
	"Received: from [208.64.202.21] (helo=example.com)\r\n" +
	"    by smtp-04-tuk1.example.com with smtp (Exim 4.90_1)\r\n" +
	"    (envelope-from <aaaaaaa@example.com>)\r\n" +
	"    id 1iJ1TK-0001P5-61\r\n" +
	"    for aaaaaaa@example.com; Fri, 11 Oct 2019 13:25:10 -0700\r\n" +
	"To: aaaaaaa@example.com\r\n" +
	"From: \"whatever\" <aaaaaaa@example.com>\r\n" +
	"Reply-To: <aaaaaaa@example.com>\r\n" +
	"Errors-To: <aaaaaaa@example.com>\r\n" +
	"X-whatever-Message-Type: AAAAAAAAAAAAAAAAAAAAAAAA\r\n" +
	"Mime-Version: 1.0\r\n" +
	"Subject: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA!\r\n" +
	"Content-Type: multipart/alternative;\r\n" +
	" boundary=\"np5da0e524c5cff\"\r\n" +
	"Message-Id: <AAAAAAAAAAAAAAAAA@aaaaaaaaaaaa.example.com>\r\n" +
	"Date: Fri, 11 Oct 2019 13:25:10 -0700\r\n" +
	"\r\n"

const testHeaderString2 = "Return-Path: <aaaaaaa@example.com>\r\n" +
	"Delivered-To: aaaaaaa@example.com\r\n" +
	"Received: from localhost (localhost [127.0.0.1])\r\n" +
	"    by example.com (Postfix) with ESMTP id 6674C25DD7\r\n" +
	"    for <aaaaaaa@example.com>; Sat, 12 Oct 2019 23:27:19 +0200 (CEST)\r\n" +
	"X-Virus-Scanned: Debian amavisd-new at example.com\r\n" +
	"X-Spam-Flag: NO\r\n" +
	"X-Spam-Score: -1.698\r\n" +
	"X-Spam-Level:\r\n" +
	"X-Spam-Status: No, score=-1.698 tagged_above=-9999 required=5\r\n" +
	"    tests=[BAYES_00=-1.9, DKIM_INVALID=0.1, DKIM_SIGNED=0.1,\r\n" +
	"    HTML_FONT_LOW_CONTRAST=0.001, HTML_MESSAGE=0.001, SPF_HELO_NONE=0.001,\r\n" +
	"    SPF_PASS=-0.001] autolearn=no autolearn_force=no\r\n" +
	"Received: from aaaaa.example.com ([127.0.0.1])\r\n" +
	"    by localhost (example.com [127.0.0.1]) (amavisd-new, port 10024)\r\n" +
	"    with ESMTP id rreNxtBzkKKg for <aaaaaaa@example.com>;\r\n" +
	"    Sat, 12 Oct 2019 23:27:17 +0200 (CEST)\r\n" +
	"Received: from aaaaaaaaa.example.com (aaaaaaaaa.example.com [52.21.114.224])\r\n" +
	"    by example.com (Postfix) with ESMTPS id 04078252B0\r\n" +
	"    for <aaaaaaa@example.com>; Sat, 12 Oct 2019 23:27:16 +0200 (CEST)\r\n" +
	"Authentication-Results: example.com;\r\n" +
	"    dkim=fail reason=\"signature verification failed\" (2048-bit key; unprotected) header.d=example.com header.i=@example.com header.b=\"dZdXFfdO\";\r\n" +
	"    dkim-atps=neutral\r\n" +
	"Received:\r\n" +
	"    by pigeon-at-10005 (OpenSMTPD) with ESMTP id 8c8bc170\r\n" +
	"    for <aaaaaaa@example.com>;\r\n" +
	"    Sat, 12 Oct 2019 21:27:15 +0000 (UTC)\r\n" +
	"DKIM-Signature: v=1; a=rsa-sha1; c=relaxed/relaxed; d=example.com; h=\r\n" +
	"    content-type:mime-version:from:to:subject:list-unsubscribe:date\r\n" +
	"    :message-id; s=pigeon; bh=KSoP6Qz2pPwRYI9o8UOCFcgfqoA=; b=dZdXFf\r\n" +
	"    dOFuwK7RaGOspcDR+26a2iQRLO7WuXahe+X/deW0tvmoaRGyF18ei3nwM7lZdHyZ\r\n" +
	"    gztpqTsZtYHfyhqf7lMplMt4uGoxo1iofM4GFRiJy2A+umfOnLRYcAb5Hulyn83c\r\n" +
	"    YldmKy0Cmy4B1uRYozyv6doMScYaIiB9STNnaaJh3oaApx4wrk5r0kav2CGc7e0/\r\n" +
	"    rLij61X8nyLFOnPzHNi1ByXsAZWxZtOe7H7mzF+Xh/WQ6y6SQnkksg+AhsJGQ1b/\r\n" +
	"    usdLollf47EJaYEuHOMKsrvIqRCWACq7Mhzu3KMi81Tl0TGhk4KfSp+6ldSIc/39\r\n" +
	"    0z8EFXV5TSoJA0dg==\r\n" +
	"Content-Type: multipart/alternative;\r\n" +
	" boundary=\"===============7074720964622344361==\"\r\n" +
	"Mime-Version: 1.0\r\n" +
	"From: example <aaaaaaaaaaaaaaaaaaaaaaaa@example.com>\r\n" +
	"To: aaaaaaa@example.com\r\n" +
	"Subject: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=\r\n" +
	" =?utf-8?q?aaaaaaaaaaaaaaaaaaaaaaaaaaaaa?=\r\n" +
	"List-Unsubscribe: <http://www.example.com/email_optout/qemail_unsubscribe?email_track_id=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX>\r\n" +
	"X-CID: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	"Date: Sat, 12 Oct 2019 21:27:15 +0000\r\n" +
	"Message-ID: <aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@aaaaaaaa.example.com>\r\n" +
	"X-SMTPAPI: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	"X-QMSG: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	" aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n" +
	"\r\n"

const testHeaderString3 = "Return-Path: <aaaaaaaaaaa@example.com>\r\n" +
	"Delivered-To: aaaaaaa@example.com\r\n" +
	"Received: from localhost (localhost [127.0.0.1])\r\n" +
	"   by example.com (Postfix) with ESMTP id C263620992\r\n" +
	"   for <aaaaaaa@example.com>; Sun, 22 Sep 2019 14:16:09 +0200 (CEST)\r\n" +
	"X-Virus-Scanned: Debian amavisd-new at example.com\r\n" +
	"X-Spam-Flag: NO\r\n" +
	"X-Spam-Score: -1.9\r\n" +
	"X-Spam-Level:\r\n" +
	"X-Spam-Status: No, score=-1.9 tagged_above=-9999 required=5\r\n" +
	"   tests=[BAYES_00=-1.9, SPF_HELO_NONE=0.001, SPF_PASS=-0.001]\r\n" +
	"   autolearn=ham autolearn_force=no\r\n" +
	"Received: from aaaaa.example.com ([127.0.0.1])\r\n" +
	"    by localhost (example.com [127.0.0.1]) (amavisd-new, port 10024)\r\n" +
	"    with ESMTP id HwPdIXhwQogO for <aaaaaaa@example.com>;\r\n" +
	"    Sun, 22 Sep 2019 14:16:08 +0200 (CEST)\r\n" +
	"Received-SPF: Pass (mailfrom) identity=mailfrom; client-ip=XXX.XXX.XXX.XXX; helo=aaaaaaaaa.example.com; envelope-from=aaaaaaaaaaa@example.com; receiver=<UNKNOWN> \r\n" +
	"Received: from aaaaaaaaa.example.com (aaaaaaaaaa.example.com [XXX.XXX.XXX.XXX])\r\n" +
	"    by example.com (Postfix) with ESMTPS id 7C70420990\r\n" +
	"    for <aaaaaaa@example.com>; Sun, 22 Sep 2019 14:16:07 +0200 (CEST)\r\n" +
	"X-Note: This Email was scanned by AppRiver SecureTide\r\n" +
	"X-Note-AR-ScanTimeLocal: 09/22/2019 8:01:04 AM\r\n" +
	"X-Note: SecureTide Build: 9/5/2019 3:33:32 PM UTC (2.8.5.0)\r\n" +
	"X-Note: Filtered by 10.246.0.223\r\n" +
	"X-Note-AR-Scan: None - PIPE\r\n" +
	"Received: by aaaaaaaaa.example.com (CommuniGate Pro PIPE 6.2.4)\r\n" +
	"    with PIPE id 12280785; Sun, 22 Sep 2019 08:01:04 -0400\r\n" +
	"Received: from [XXX.XXX.XXX.XXX] (HELO aaaaaaaaaaaa.example.com)\r\n" +
	"    by aaaaaaaaa.example.com (CommuniGate Pro SMTP 6.2.4)\r\n" +
	"    with ESMTP id 12280776 for aaaaaaa@example.com; Sun, 22 Sep 2019 08:01:00 -0400\r\n" +
	"Received: from aaaaaaaaaaaaaaaaaaaaaaaaaa.local (XXX.XXX.XXX.XXX) by\r\n" +
	"    aaaaaaaaaaaaaaaaaaaaaaaaaa.local (XXX.XXX.XXX.XXX) with Microsoft SMTP Server\r\n" +
	"    (version=TLS1_2, cipher=TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256_P256) id\r\n" +
	"    X.X.XXXX.X; Sun, 22 Sep 2019 08:01:00 -0400\r\n" +
	"Received: from aaaaaaaaaaaaaaaaaaaaaaaaaa.local ([XXX.XXX.XXX.XXX]) by\r\n" +
	"    aaaaaaaaaaaaaaaaaaaaaaaaaa.local ([XXX.XXX.XXX.XXX]) with mapi id\r\n" +
	"    15.01.1779.000; Sun, 22 Sep 2019 08:01:00 -0400\r\n" +
	"From: \"AAAAAAAAAAAAAA\" <aaaaaaaaaaa@example.com>\r\n" +
	"To: AAAAAAAAAAA <aaaaaaa@example.com>\r\n" +
	"Subject: RE: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\n" +
	"    AAAAAAAAAAAAAAA\r\n" +
	"Thread-Topic: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\n" +
	"    AAAAAAAAAAAAAAA\r\n" +
	"Thread-Index: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\r\n" +
	"Date: Sun, 22 Sep 2019 12:01:00 +0000\r\n" +
	"Message-ID: <aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com>\r\n" +
	"References: <aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com>\r\n" +
	"    <aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com>\r\n" +
	"In-Reply-To: <aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com>\r\n" +
	"Accept-Language: en-US\r\n" +
	"Content-Language: en-US\r\n" +
	"X-MS-Has-Attach:\r\n" +
	"X-MS-TNEF-Correlator:\r\n" +
	"x-rerouted-by-exchange:\r\n" +
	"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"Mime-Version: 1.0\r\n" +
	"X-Note: This Email was scanned by AppRiver SecureTide\r\n" +
	"X-Note-AR-ScanTimeLocal: 09/22/2019 8:01:00 AM\r\n" +
	"X-Note: SecureTide Build: 9/5/2019 3:33:32 PM UTC (2.8.5.0)\r\n" +
	"X-Note: Filtered by XXX.XXX.XXX.XXX\r\n" +
	"X-Policy: example.com\r\n" +
	"X-Primary: example.com@example.com\r\n" +
	"X-Note-Sender:  <aaaaaaaaaaa@example.com>\r\n" +
	"X-Virus-Scan: V-\r\n" +
	"X-Note-SnifferID: 0\r\n" +
	"X-GBUdb-Analysis: 1, XXX.XXX.XXX.XXX, Ugly c=0.765185 p=-0.992849 Source White\r\n" +
	"X-Signature-Violations:\r\n" +
	"    0-0-0-2154-c\r\n" +
	"X-Note-419: 0 ms. Fail:1 Chk:1354 of 1354 total\r\n" +
	"X-Note: VSCH-CT/SI: 1-1354/SG:1 9/22/2019 8:00:50 AM\r\n" +
	"X-Note: Spam Tests Failed: \r\n" +
	"X-Country-Path: PRIVATE->PRIVATE->\r\n" +
	"X-Note-Sending-IP: XXX.XXX.XXX.XXX\r\n" +
	"X-Note-Reverse-DNS: \r\n" +
	"X-Note-Return-Path: shahid.shah@example.com\r\n" +
	"X-Note: User Rule Hits: \r\n" +
	"X-Note: Global Rule Hits: G694 G695 G696 G697 G715 G716 G717 G870 \r\n" +
	"X-Note: Encrypt Rule Hits: \r\n" +
	"X-Note: Mail Class: VALID\r\n" +
	"X-Note-ECS-IP:\r\n" +
	"X-Note-ECS-Recip: aaaaaaa@example.com\r\n" +
	"\r\n"

// Assign to global variable to prevent compiler optimizations.
var hdr Header

func BenchmarkTextprotoReadHeader(b *testing.B) {
	bench := func(name, blob string) {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			r := bufio.NewReader(strings.NewReader(blob))

			var err error
			for i := 0; i < b.N; i++ {
				r.Reset(strings.NewReader(blob))
				hdr, err = ReadHeader(r, nil)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	bench("1", testHeaderString1)
	bench("2", testHeaderString2)
	bench("3", testHeaderString3)
}

var mimeHdr textproto.MIMEHeader

func BenchmarkStdlibReadHeader(b *testing.B) {
	bench := func(name, blob string) {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			r := bufio.NewReader(strings.NewReader(blob))
			tr := textproto.NewReader(r)

			var err error
			for i := 0; i < b.N; i++ {
				r.Reset(strings.NewReader(blob))
				mimeHdr, err = tr.ReadMIMEHeader()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	bench("1", testHeaderString1)
	bench("2", testHeaderString2)
	bench("3", testHeaderString3)
}
